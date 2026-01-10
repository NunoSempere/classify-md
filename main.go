package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Topic struct {
	name     string
	heading  string   // Store the full heading including # symbols
	keywords []string
}

func main() {
	automatic := flag.Bool("a", false, "Enable fully automatic classification using OpenAI")
	flag.BoolVar(automatic, "automatic", false, "Enable fully automatic classification using OpenAI")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fmt.Println("Usage: classify [-a|--automatic] <topics_file> <markdown_file>")
		fmt.Println("  -a, --automatic  Enable fully automatic classification using OpenAI")
		fmt.Println("                   Requires OPENAI_API_KEY environment variable")
		os.Exit(1)
	}

	topicsFile := args[0]
	markdownFile := args[1]

	var openaiToken string
	if *automatic {
		openaiToken = os.Getenv("OPENAI_API_KEY")
		if openaiToken == "" {
			fmt.Println("Error: OPENAI_API_KEY environment variable is required for automatic mode")
			os.Exit(1)
		}
	}

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

	// Determine output file name
	fileExt := ".md" // Default extension
	baseName := markdownFile
	if lastDot := strings.LastIndex(markdownFile, "."); lastDot >= 0 {
		fileExt = markdownFile[lastDot:]
		baseName = markdownFile[:lastDot]
	}
	outputFile := baseName + ".ordered" + fileExt

	// Read existing content if the file exists
	existingContent, err := readExistingOrderedFile(outputFile)
	if err != nil {
		fmt.Printf("Error reading existing ordered file: %v\n", err)
		os.Exit(1)
	}

	// Create a map to store sections for each topic
	topicContent := make(map[string][]string)
	for _, topic := range topics {
		// Initialize with existing content if any
		if existing, ok := existingContent[topic.name]; ok {
			topicContent[topic.name] = existing
		} else {
			topicContent[topic.name] = []string{}
		}
	}

	// Create a reader for user input
	reader := bufio.NewReader(os.Stdin)

	// Process each section
	for _, section := range sections {
		fmt.Println("\nSection content:")
		fmt.Println("-------------------")
		fmt.Println(section)
		fmt.Println("-------------------")
		
		// Check if section already exists in any topic
		sectionFound := false
		for _, topic := range topics {
			if sectionExists(section, topicContent[topic.name]) {
				fmt.Printf("\nSection already exists under topic '%s'\n", topic.name)
				sectionFound = true
				break
			}
		}
		
		if sectionFound {
			continue
		}

		// Automatic mode: use OpenAI for classification
		if *automatic {
			topicNames := make([]string, len(topics))
			for i, t := range topics {
				topicNames[i] = t.name
			}

			topicIndex, reasoning, err := ClassifySection(section, topicNames, openaiToken)
			if err != nil {
				fmt.Printf("\nOpenAI classification error: %v\n", err)
				fmt.Println("Skipping section.")
				continue
			}

			selectedTopic := topics[topicIndex].name
			fmt.Printf("\nClassified as: '%s'\n", selectedTopic)
			fmt.Printf("Reasoning: %s\n", reasoning)
			topicContent[selectedTopic] = append(topicContent[selectedTopic], section)
			continue
		}

		// Manual mode: check if topic name or any keywords appear in the section
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
			fmt.Printf("\nDefault classification: '%s' (matched topic name or keyword)\n", matchedTopic)
			fmt.Print("Accept this classification? [y/n]: ")
			
			acceptInput, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input. Skipping section.")
				continue
			}
			
			acceptInput = strings.TrimSpace(strings.ToLower(acceptInput))
			
			if acceptInput == "y" || acceptInput == "yes" {
				topicContent[matchedTopic] = append(topicContent[matchedTopic], section)
				continue
			}
			
			// User declined - fall through to manual selection
			fmt.Println("\nPlease select a different topic:")
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
				newTopic := Topic{name: topicName}
				topics = append(topics, newTopic)
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
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Write all content to the file
	for _, topic := range topics {
		// Write topic header with original heading symbols
		_, err := file.WriteString(fmt.Sprintf("%s\n\n", topic.heading))
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

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		fmt.Printf("\nCreated new file: %s\n", outputFile)
	} else {
		fmt.Printf("\nUpdated existing file: %s\n", outputFile)
	}
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
		heading := parts[0]
		name := strings.TrimSpace(strings.TrimLeft(heading, "#")) // Remove # symbols for the name
		
		topic := Topic{
			name:    name,
			heading: heading,
		}
		
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

func readExistingOrderedFile(filename string) (map[string][]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string][]string), nil
		}
		return nil, err
	}

	existingContent := make(map[string][]string)
	var currentTopic string
	var currentSection strings.Builder
	
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := scanner.Text()
		
		// Check if this is a topic line (starts with any number of # symbols)
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			// Save previous section if it exists
			if currentSection.Len() > 0 && currentTopic != "" {
				section := strings.TrimSpace(currentSection.String())
				if section != "" {
					existingContent[currentTopic] = append(existingContent[currentTopic], section)
				}
				currentSection.Reset()
			}
			// Set new topic (remove # symbols)
			currentTopic = strings.TrimSpace(strings.TrimLeft(line, "#"))
		} else if strings.TrimSpace(line) != "" && currentTopic != "" {
			if currentSection.Len() > 0 {
				currentSection.WriteString("\n")
			}
			currentSection.WriteString(line)
		} else if strings.TrimSpace(line) == "" && currentSection.Len() > 0 {
			// Empty line - save current section if we have one
			section := strings.TrimSpace(currentSection.String())
			if section != "" && currentTopic != "" {
				existingContent[currentTopic] = append(existingContent[currentTopic], section)
			}
			currentSection.Reset()
		}
	}

	// Add final section if it exists
	if currentSection.Len() > 0 && currentTopic != "" {
		section := strings.TrimSpace(currentSection.String())
		if section != "" {
			existingContent[currentTopic] = append(existingContent[currentTopic], section)
		}
	}

	return existingContent, scanner.Err()
}

func sectionExists(section string, existingSections []string) bool {
	for _, existing := range existingSections {
		if strings.TrimSpace(existing) == strings.TrimSpace(section) {
			return true
		}
	}
	return false
}
