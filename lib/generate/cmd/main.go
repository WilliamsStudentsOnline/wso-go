package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jinzhu/inflection"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "WSO Generator"
	app.HideVersion = true
	app.Usage = "a scaffolding generator generator"
	app.HelpName = "generate"

	app.Commands = []cli.Command{
		{
			Name:      "service",
			Usage:     "generate a service",
			Action:    generateService,
			ArgsUsage: "[service_name]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "model,m",
					Usage: "a model that the service depends on",
				},
			},
		},
		{
			Name:      "model",
			Usage:     "generate a model",
			Action:    generateModel,
			ArgsUsage: "[ModelName]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "table,t",
					Usage: "the model's table",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println("Error:", err.Error())
	}

}

func generateModel(c *cli.Context) error {
	name := c.Args().First()

	if name == "" {
		return errors.New("a name is required")
	}

	name = strings.Title(name)
	nameLC := strings.ToLower(name)

	table := strings.ToLower(c.String("table"))
	if table == "" {
		table = inflection.Plural(nameLC)
	}

	path, err := filepath.Abs("models")
	if err != nil {
		return err
	}

	// Schema path
	schemaPath := filepath.Join(path, nameLC+"_schema.go")

	// Template the schema file
	schemTmpl, err := template.New("schema").Parse(modelSchemaTmpl)
	if err != nil {
		return err
	}

	schemaFile, err := os.Create(schemaPath)
	if err != nil {
		return err
	}
	defer schemaFile.Close()

	err = schemTmpl.Execute(schemaFile, map[string]interface{}{
		"name":  name,
		"table": table,
	})
	if err != nil {
		return err
	}

	// Model path
	modelPath := filepath.Join(path, nameLC+".go")

	// Template the model file
	modTmpl, err := template.New("model").Parse(modelModelTmpl)
	if err != nil {
		return err
	}

	modelFile, err := os.Create(modelPath)
	if err != nil {
		return err
	}
	defer modelFile.Close()

	err = modTmpl.Execute(modelFile, map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return err
	}

	log.Println("Schema generated at", schemaPath)
	log.Println("Model generated at", modelPath)
	log.Println("Please edit these files with new data")

	return nil
}

func generateService(c *cli.Context) error {
	name := c.Args().First()

	if name == "" {
		return errors.New("a name is required")
	}

	model := c.String("model")
	if model == "" {
		// Convert from underscore_case to CamelCase
		model = strings.ReplaceAll(strings.Title(strings.ReplaceAll(name, "_", " ")), " ", "")
	}

	path, err := filepath.Abs(filepath.Join("services", name))
	if err != nil {
		return err
	}

	err = os.Mkdir(path, os.ModePerm)
	if err != nil {
		return err
	}

	// Controller path
	controllerPath := filepath.Join(path, "controller.go")

	// Template the controller file
	ctrlTmpl, err := template.New("controller").Parse(serviceControllerTmpl)
	if err != nil {
		return err
	}

	ctrlFile, err := os.Create(controllerPath)
	if err != nil {
		return err
	}
	defer ctrlFile.Close()

	err = ctrlTmpl.Execute(ctrlFile, map[string]interface{}{
		"name":    name,
		"model":   model,
		"modelDC": strings.ToLower(model[:1]) + model[1:],
	})
	if err != nil {
		return err
	}

	// Router path
	routerPath := filepath.Join(path, "router.go")

	// Template the router file
	rtrTmpl, err := template.New("router").Parse(serviceRouterTmpl)
	if err != nil {
		return err
	}

	rtrFile, err := os.Create(routerPath)
	if err != nil {
		return err
	}
	defer rtrFile.Close()

	err = rtrTmpl.Execute(rtrFile, map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return err
	}

	log.Println("Service generated at", path)
	log.Println("Please edit these files with new data")

	return nil
}

var serviceControllerTmpl = `package {{.name}}

import (
	"errors"
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Controller struct {
	services.BaseController
	// Put a model here, like:
	{{.modelDC}}Model *models.{{.model}}
}

// Construct a new user controller
func NewController(db *gorm.DB) *Controller {
	return &Controller{
		{{.modelDC}}Model: &models.{{.model}}{
			BaseModel: models.BaseModel{
				DB: db,
			},
		},
	}
}

// Endpoint example
/*
func (t *Controller) FetchAllUsers(c *gin.Context) {
	var users []models.User
	err := t.userModel.GetAllUsers(&users)
	
	if err != nil {
		t.RespondErrorCode(http.StatusInternalServerError, err, c)
		return
	}

	t.RespondOK(users, c)
}
*/
`

var serviceRouterTmpl = `package {{.name}}

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetupRouter(r gin.IRouter, db *gorm.DB) {
	c := NewController(db)
	// Example route:
	/*
	r.GET("/", c.FetchAllUsers)
	*/
}
`

var modelSchemaTmpl = `package models

import "github.com/jinzhu/gorm"

// {{.name}} Schema
type {{.name}} struct {
	BaseSchema
}

func (*{{.name}}) TableName() string {
	return "{{.table}}"
}
`

var modelModelTmpl = `package models

// {{.name}} Model
type {{.name}}Model struct {
	*BaseModel
}

func New{{.name}}Model(db *gorm.DB) *{{.name}}Model {
	return &{{.name}}Model{
		BaseModel: NewBaseModel(db),
	}
}
`
