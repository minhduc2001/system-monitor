package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"go-runner/internal/config"
	"go-runner/internal/db"
	"go-runner/internal/project"
)

type ExampleData struct {
	ProjectGroups []project.CreateProjectGroupRequest `json:"project_groups"`
	Microservices []project.CreateProjectRequest      `json:"microservices"`
}

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	database := db.InitDB(cfg)

	// Read example data
	data, err := ioutil.ReadFile("examples/microservices.json")
	if err != nil {
		log.Fatalf("Error reading example data: %v", err)
	}

	var exampleData ExampleData
	if err := json.Unmarshal(data, &exampleData); err != nil {
		log.Fatalf("Error parsing example data: %v", err)
	}

	// Create project groups
	fmt.Println("Creating project groups...")
	for _, groupReq := range exampleData.ProjectGroups {
		group := project.ProjectGroup{
			Name:        groupReq.Name,
			Description: groupReq.Description,
			Color:       groupReq.Color,
		}
		if err := database.Create(&group).Error; err != nil {
			log.Printf("Error creating group %s: %v", groupReq.Name, err)
		} else {
			fmt.Printf("✅ Created group: %s\n", group.Name)
		}
	}

	// Create microservices
	fmt.Println("\nCreating microservices...")
	for _, serviceReq := range exampleData.Microservices {
		service := project.Project{
			Name:           serviceReq.Name,
			Description:    serviceReq.Description,
			Type:           serviceReq.Type,
			GroupID:        serviceReq.GroupID,
			Path:           serviceReq.Path,
			Command:        serviceReq.Command,
			Args:           serviceReq.Args,
			WorkingDir:     serviceReq.WorkingDir,
			Port:           serviceReq.Port,
			Ports:          serviceReq.Ports,
			Environment:    serviceReq.Environment,
			EnvFile:        serviceReq.EnvFile,
			EnvVars:        serviceReq.EnvVars,
			Editor:         serviceReq.Editor,
			EditorArgs:     serviceReq.EditorArgs,
			HealthCheckURL: serviceReq.HealthCheckURL,
			AutoRestart:    serviceReq.AutoRestart,
			MaxRestarts:    serviceReq.MaxRestarts,
			CPULimit:       serviceReq.CPULimit,
			MemoryLimit:    serviceReq.MemoryLimit,
		}
		if err := database.Create(&service).Error; err != nil {
			log.Printf("Error creating service %s: %v", serviceReq.Name, err)
		} else {
			fmt.Printf("✅ Created service: %s (%s)\n", service.Name, service.Type)
		}
	}

	fmt.Println("\n🎉 Example data imported successfully!")
	fmt.Println("You can now start the server and test the API endpoints.")
}
