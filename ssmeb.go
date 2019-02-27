package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	ssm "github.com/aws/aws-sdk-go/service/ssm"
	yaml "gopkg.in/yaml.v2"
)

// parameters is the type this program expects its input file to have.
type parameters struct {
	// Component holds parameters owned by this app
	Component []parameter `yaml:"component"`
	// External holds parameters external to this app. They can't be set.
	External []parameter `yaml:"external"`
}

// parameter holds info about an ssm parameter
type parameter struct {
	// Name is the name of environment variable stored in the ssm parameter (required)
	Name string `yaml:"option_name"`
	// Description is an optional string describing parameter
	Description string `yaml:"description"`
	// Path is key for the parameter on the Systems Manager
	Path string `yaml:"path"`
	// Value is the value stored on the Systems Manager. This is optional but useful when using the set mode
	Value string `yaml:"value"`
}

// ebOptionSettings is the output format of this program, which conforms with
// the format used for elastic beanstalk extensions
type ebOptionSettings struct {
	Options []ebOption `yaml:"option_settings"`
}

// ebOption hold info about a beanstalk option
type ebOption struct {
	// Name is the option name
	Name string `yaml:"option_name"`
	// Value is the option value
	Value string `yaml:"value"`
}

func main() {

	// read and parse input
	var input string
	flag.StringVar(&input, "input", "", "input template environment variables config")
	flag.StringVar(&input, "i", "", "`input` flag shorthand")

	var output string
	flag.StringVar(&output, "output", "", "destination of the resulting elastic beanstalk data")
	flag.StringVar(&output, "o", "", "`output` flag shorthand")

	var environment string
	flag.StringVar(&environment, "environment", "", "environment name used as prefix for the ssm parameters (e.g. codacy)")
	flag.StringVar(&environment, "e", "", "`environment` flag shorthand")

	var mode string
	flag.StringVar(&mode, "mode", "get", "enable set or get mode")
	flag.StringVar(&mode, "m", "get", "`mode` flag shorthand")

	flag.Parse()

	fmt.Fprintln(os.Stderr, "-----------------------------------------")
	fmt.Fprintln(os.Stderr, "input:       ", input)
	fmt.Fprintln(os.Stderr, "output:      ", output)
	fmt.Fprintln(os.Stderr, "environment: ", environment)
	fmt.Fprintln(os.Stderr, "mode:        ", mode)
	fmt.Fprintln(os.Stderr, "-----------------------------------------")

	if input == "" {
		log.Fatal("Missing mandatory argument: `input`")
	}
	parameters, err := readParametersFile(input, environment)
	if err != nil {
		log.Fatalf("Error reading file `%s`: %v", input, err)
	}

	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	if mode == "get" {
		ebOptions, err := getBeanstalkOptions(session, parameters)
		if err != nil {
			log.Fatalf("Error getting values: %v", err)
		}

		ebYaml, err := yaml.Marshal(ebOptions)
		if err != nil {
			log.Fatalf("Error marshaling beanstalk options: %v", err)
		}
		if output == "" {
			fmt.Println(string(ebYaml))
		} else {
			err = writeToFile(output, ebYaml)
			if err != nil {
				log.Fatalf("Error writing to file `%s`", output)
			}
		}
	} else if mode == "set" {
		err := setBeanstalkOptions(session, parameters)
		if err != nil {
			log.Fatalf("Error setting values: %v", err)
		}
	} else {
		log.Fatalf("Invalid mode: %s", mode)
	}

}

// readParametersFile reads parameter from a file with name filename, and prepends `/environment`
// to its path if the environment is not an empty string
func readParametersFile(filename string, environment string) (parameters, error) {
	var parameters parameters
	inputFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return parameters, err
	}

	err = yaml.Unmarshal(inputFile, &parameters)
	if err != nil {
		return parameters, err
	}

	if environment != "" {
		for i, par := range parameters.Component {
			parameters.Component[i].Path = "/" + environment + par.Path
		}
		for i, par := range parameters.External {
			parameters.External[i].Path = "/" + environment + par.Path
		}
	}

	return parameters, nil
}

// getBeanstalkOptions converts the parameters into ebOptionSettings, by getting the data
// for each one from SSM using a client created from the provided session.
func getBeanstalkOptions(session *session.Session, parameters parameters) (ebOptionSettings, error) {
	ssmClient := ssm.New(session)

	var eb ebOptionSettings

	for _, par := range parameters.Component {
		fmt.Fprintf(os.Stderr, "* Getting `%s` from path `%s`... ", par.Name, par.Path)
		parOutput, err := ssmClient.GetParameter(&ssm.GetParameterInput{Name: &par.Path})
		if err != nil {
			return eb, err
		}
		eb.Options = append(eb.Options, ebOption{Name: par.Name, Value: *parOutput.Parameter.Value})
		fmt.Fprintln(os.Stderr, "OK")
	}

	for _, par := range parameters.External {
		fmt.Fprintf(os.Stderr, "* Getting `%s` from path `%s`... ", par.Name, par.Path)
		parOutput, err := ssmClient.GetParameter(&ssm.GetParameterInput{Name: &par.Path})
		if err != nil {
			return eb, err
		}
		eb.Options = append(eb.Options, ebOption{Name: par.Name, Value: *parOutput.Parameter.Value})
		fmt.Fprintln(os.Stderr, "OK")
	}

	return eb, nil
}

// setBeanstalkOptions sends parameters into SSM using a client created from the provided session
func setBeanstalkOptions(session *session.Session, parameters parameters) error {
	ssmClient := ssm.New(session)

	for _, par := range parameters.Component {

		value := par.Value
		if value == "" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Printf("* Input value for `%s`: ", par.Path)

			text, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			value = strings.Replace(text, "\n", "", -1)
		} else {
			fmt.Printf("* Setting value for `%s`...\n", par.Path)
		}

		overwrite := true
		parType := "String"
		ssmPar := ssm.PutParameterInput{
			Name:        &par.Path,
			Description: &par.Description,
			Value:       &value,
			Overwrite:   &overwrite,
			Type:        &parType,
		}
		fmt.Println(ssmPar)
		putOutput, err := ssmClient.PutParameter(&ssmPar)
		if err != nil {
			return err
		}
		fmt.Println(putOutput)
	}
	return nil
}

// writeToFile saves the data to a file whose name is given in output
func writeToFile(output string, data []byte) error {
	outFile, err := os.Create(output)
	if err != nil {
		return err
	}
	bytesOut, err := outFile.Write(data)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%d bytes written successfully to `%s`\n", bytesOut, output)
	err = outFile.Close()
	if err != nil {
		return err
	}
	return nil
}
