package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// Check if swag is installed
	_, err := exec.LookPath("swag")
	if err != nil {
		fmt.Println("Error: swag command not found. Please install it with:")
		fmt.Println("go install github.com/swaggo/swag/cmd/swag@latest")
		os.Exit(1)
	}

	// Step 1: Generate Swagger documentation
	fmt.Println("Generating Swagger documentation...")
	swagCmd := exec.Command("swag", "init", "-g", "cmd/server/main.go", "-d", "./")
	swagCmd.Stdout = os.Stdout
	swagCmd.Stderr = os.Stderr

	err = swagCmd.Run()
	if err != nil {
		fmt.Println("Error generating Swagger documentation:", err)
		os.Exit(1)
	}
	fmt.Println("Swagger documentation generated successfully!")

	// Step 2: Run the application
	fmt.Println("Starting the application...")
	appCmd := exec.Command("go", "run", "cmd/server/main.go")
	appCmd.Stdout = os.Stdout
	appCmd.Stderr = os.Stderr

	err = appCmd.Run()
	if err != nil {
		fmt.Println("Error running application:", err)
		os.Exit(1)
	}
}
