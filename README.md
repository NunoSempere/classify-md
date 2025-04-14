# Topic Classifier

A simple tool to classify sections of a markdown file into separate topic files.

## Usage

```bash
go run main.go <topics_file> <markdown_file>
```

The program will:

1. Read the topics from the topics file (one topic per line)
2. Read the markdown file and split it into sections (separated by empty lines)
3. For each section:
   - Display the section content
   - Show available topics
   - Prompt you to choose a topic
4. Create a single output file with ".ordered" extension containing all sections organized by topics

## Example

```bash
go run main.go example/topics.txt example/markdown.md
```

This will create `example_markdown.ordered.md` containing all sections organized under their respective topic headers.

## Output Format

The output file will be structured like this:

```markdown
# Topic1

Section content...
Another section content...

# Topic2

Section content...

# Topic3

Section content...
```

## To do

- [ ] Add default topics / subtopics
- [ ] Be able to use regexes?
- [ ] Add topic
- [ ] Change extension
- [ ] Merge two files with the same topics
- [ ] Share with forecasters
