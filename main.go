package main

import (
	"fmt"
	"log"
)

func main() {
	input := "ALL_DATASET_UNCLEAN_11_08.csv"
	output := "label_transformed_output.csv"
	if err := TransformLabelCSV(input, output); err != nil {
		log.Fatalf("Transformation failed: %v", err)
	}
	fmt.Println("Label Transformation succeeded!")
}
