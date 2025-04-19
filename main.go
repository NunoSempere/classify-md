package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Topic struct {
	name     string
	keywords []string
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: topic_classifier <topics_file> <markdown_file>")
		os.Exit(1)
	}

	topicsFile := os.Args[1]
	markdownFile := os.Args[2]

	topics, err := readTopics(topicsFile)
	if err != nil {
		fmt.Printf("Error reading topics file: %v\n", err)
		os.Exit(1)
	}

	sections, err := readMarkdownSections(markdownFile)
	if err != nil {
		fmt.Printf("Error reading markdown file: %v\n", err)
		os.Exit(1)
	}

	// Create a map to store sections for each topic
	topicContent := make(map[string][]string)
	for _, topic := range topics {
		topicContent[topic.name] = []string{}
	}

	// Create a reader for user input
	reader := bufio.NewReader(os.Stdin)

	// Process each section
	for _, section := range sections {
		fmt.Println("\nSection content:")
		fmt.Println("-------------------")
		fmt.Println(section)
		fmt.Println("-------------------")
		
		// Check if topic name or any keywords appear in the section
		foundTopic := false
		var matchedTopic string
		sectionLower := strings.ToLower(section)
		
		for _, topic := range topics {
			// Check topic name
			if strings.Contains(sectionLower, strings.ToLower(topic.name)) {
				foundTopic = true
				matchedTopic = topic.name
				break
			}
			
			// Check keywords
			for _, keyword := range topic.keywords {
				if strings.Contains(sectionLower, keyword) {
					foundTopic = true
					matchedTopic = topic.name
					break
				}
			}
			if foundTopic {
				break
			}
		}

		if foundTopic {
			fmt.Printf("\nAutomatically classified as '%s' (matched topic name or keyword)\n", matchedTopic)
			topicContent[matchedTopic] = append(topicContent[matchedTopic], section)
			continue
		}
		
		fmt.Println("\nAvailable topics:")
		for i, topic := range topics {
			fmt.Printf("%d. %s\n", i+1, topic.name)
		}
		fmt.Println("\nPress 'a' to add a new topic")

		var choice int
		for {
			fmt.Print("\nEnter the number of the corresponding topic (or 'a' to add): ")
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input. Please try again.")
				continue
			}
			
			input = strings.TrimSpace(input)
			
			if input == "a" {
				fmt.Print("Enter new topic name: ")
				topicName, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Error reading topic name. Please try again.")
					continue
				}
				topicName = strings.TrimSpace(topicName)
				if topicName == "" {
					fmt.Println("Topic name cannot be empty. Please try again.")
					continue
				}
				
				// Add the new topic
				topics = append(topics, Topic{name: topicName})
				topicContent[topicName] = []string{}
				fmt.Printf("Added new topic: %s\n", topicName)
				
				// Show updated topic list
				fmt.Println("\nAvailable topics:")
				for i, topic := range topics {
					fmt.Printf("%d. %s\n", i+1, topic.name)
				}
				continue
			}
			
			// Convert input to integer
			choice, err = strconv.Atoi(input)
			if err == nil && choice > 0 && choice <= len(topics) {
				break
			}
			fmt.Println("Invalid choice. Please enter a number between 1 and", len(topics), "or 'a' to add a topic")
		}

		// Add the section to the chosen topic
		selectedTopic := topics[choice-1].name
		topicContent[selectedTopic] = append(topicContent[selectedTopic], section)
	}

	// Create the output file
	fileExt := ".md" // Default extension
	baseName := markdownFile
	if lastDot := strings.LastIndex(markdownFile, "."); lastDot >= 0 {
		fileExt = markdownFile[lastDot:]
		baseName = markdownFile[:lastDot]
	}
	outputFile := baseName + ".ordered" + fileExt
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Write all content to the file
	for _, topic := range topics {
		// Write topic header
		_, err := file.WriteString(fmt.Sprintf("%s\n\n", topic.name))
		if err != nil {
			fmt.Printf("Error writing topic header: %v\n", err)
			os.Exit(1)
		}

		// Write all sections for this topic
		for _, section := range topicContent[topic.name] {
			_, err := file.WriteString(section + "\n\n")
			if err != nil {
				fmt.Printf("Error writing section: %v\n", err)
				continue
			}
		}
	}

	fmt.Printf("\nClassified sections have been saved to: %s\n", outputFile)
}

func readTopics(filename string) ([]Topic, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var topics []Topic
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Split line by ':' to separate topic name and keywords
		parts := strings.Split(line, ":")
		topic := Topic{name: strings.TrimSpace(parts[0])}
		
		// If there are keywords, process them
		if len(parts) > 1 {
			// Split keywords by comma and trim spaces
			keywords := strings.Split(parts[1], ",")
			for _, keyword := range keywords {
				keyword = strings.TrimSpace(keyword)
				if keyword != "" {
					topic.keywords = append(topic.keywords, strings.ToLower(keyword))
				}
			}
		}
		
		topics = append(topics, topic)
	}
	return topics, scanner.Err()
}

func readMarkdownSections(filename string) ([]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Split content into sections (parts separated by empty lines)
	var sections []string
	var currentSection strings.Builder
	
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	emptyLine := false

	for scanner.Scan() {
		line := scanner.Text()
		
		if strings.TrimSpace(line) == "" {
			if currentSection.Len() > 0 && !emptyLine {
				emptyLine = true
			}
		} else {
			if emptyLine && currentSection.Len() > 0 {
				sections = append(sections, strings.TrimSpace(currentSection.String()))
				currentSection.Reset()
			}
			if currentSection.Len() > 0 {
				currentSection.WriteString("\n")
			}
			currentSection.WriteString(line)
			emptyLine = false
		}
	}

	// Add the last section if it exists
	if currentSection.Len() > 0 {
		sections = append(sections, strings.TrimSpace(currentSection.String()))
	}

	return sections, scanner.Err()
}
