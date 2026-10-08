package main

import (
	"fmt"
	"regexp"
)

func main() {
	html := `
	<div class="product-gallery">
        {{#if product.image_url}}
        <img
          src="{{product.image_url}}"
          alt="{{product.name}}"
          class="main-image"
          id="mainImage"
        />
        {{else}}
        <div class="product-placeholder" id="mainImage"><span>📦</span></div>
        {{/if}} {{#if product.images}}
	</div>
	`

	fmt.Println("--- Original HTML ---")
	fmt.Println(html)

	// Simulate product.ImageURL is empty (false)
	html = processConditional(html, "product.image_url", false)

	fmt.Println("\n--- Processed HTML (Keep False) ---")
	fmt.Println(html)
}

// Copied from theme_renderer.go
func processConditional(html, condition string, keepTrue bool) string {
	// 1. Find Opening Tag
	reOpenPattern := `(?i)\{\{\s*#if\s+` + regexp.QuoteMeta(condition) + `\s*\}\}`
	reOpen := regexp.MustCompile(reOpenPattern)
	
	// General regexes for nesting check
	reAnyIf := regexp.MustCompile(`(?i)\{\{\s*#if\s+[\s\S]+?\}\}`) // Matches any {{#if ...}}
	reClose := regexp.MustCompile(`\{\{\s*/if\s*\}\}`)
	reElse := regexp.MustCompile(`\{\{\s*else\s*\}\}`)

	for {
		// Find start of our target conditional
		loc := reOpen.FindStringIndex(html)
		if loc == nil {
			break
		}
		
		startTagStart := loc[0]
		startTagEnd := loc[1]
		
		// Find matching closing tag with nesting logic
		substring := html[startTagEnd:]
		depth := 1
		var closingTagStart, closingTagEnd int
		found := false
		
		// We scan the substring relative index
		currentIndex := 0
		
		for {
			// Find next {{#if}} or {{/if}}
			locNextIf := reAnyIf.FindStringIndex(substring[currentIndex:])
			locNextClose := reClose.FindStringIndex(substring[currentIndex:])
			
			// Determine which one comes first
			var nextKey string // "if" or "close"
			var nextLoc []int
			
			// Helper to verify existence
			if locNextIf != nil && locNextClose != nil {
				if locNextIf[0] < locNextClose[0] {
					nextKey = "if"
					nextLoc = locNextIf
				} else {
					nextKey = "close"
					nextLoc = locNextClose
				}
			} else if locNextIf != nil {
				nextKey = "if"
				nextLoc = locNextIf
			} else if locNextClose != nil {
				nextKey = "close"
				nextLoc = locNextClose
			} else {
				break 
			}
			
			// Adjust loc to absolute index within substring
			absStart := currentIndex + nextLoc[0]
			absEnd := currentIndex + nextLoc[1]
			
			if nextKey == "if" {
				depth++
			} else {
				depth--
			}
			
			currentIndex = absEnd
			
			if depth == 0 {
				closingTagStart = startTagEnd + absStart
				closingTagEnd = startTagEnd + absEnd
				found = true
				break
			}
		}
		
		if !found {
			fmt.Println("ERROR: No matching closing tag found")
			return html
		}
		
		// 3. Extract Block Content
		blockContent := html[startTagEnd:closingTagStart]
		
		// 4. Handle Else (only top-level else!)
		var finalContent string
		
		// Find valid ELSE at depth 0 relative to blockContent
		elsePos := -1
		depth = 0
		curIdx := 0
		for {
			locIf := reAnyIf.FindStringIndex(blockContent[curIdx:])
			locClose := reClose.FindStringIndex(blockContent[curIdx:])
			locE := reElse.FindStringIndex(blockContent[curIdx:])
			
			// Find nearest event
			minIdx := len(blockContent) + 1
			eventType := ""
			
			if locIf != nil && locIf[0] < minIdx { minIdx = locIf[0]; eventType = "if" }
			if locClose != nil && locClose[0] < minIdx { minIdx = locClose[0]; eventType = "close" }
			if locE != nil && locE[0] < minIdx { minIdx = locE[0]; eventType = "else" }
			
			if eventType == "" { break }
			
			realStart := curIdx + minIdx
			
			if eventType == "if" {
				depth++
				curIdx = realStart + (locIf[1] - locIf[0])
			} else if eventType == "close" {
				depth--
				curIdx = realStart + (locClose[1] - locClose[0])
			} else if eventType == "else" {
				if depth == 0 {
					elsePos = realStart
					break // Found our else
				}
				curIdx = realStart + (locE[1] - locE[0])
			}
		}
		
		if elsePos != -1 {
			trueBlock := blockContent[:elsePos]
			// reElse match length
			elseMatch := reElse.FindString(blockContent[elsePos:])
			falseBlock := blockContent[elsePos+len(elseMatch):]
			
			if keepTrue {
				finalContent = trueBlock
			} else {
				finalContent = falseBlock
			}
		} else {
			if keepTrue {
				finalContent = blockContent
			} else {
				finalContent = ""
			}
		}
		
		// 5. Reconstruct HTML
		html = html[:startTagStart] + finalContent + html[closingTagEnd:]
	}
	
	return html
}
