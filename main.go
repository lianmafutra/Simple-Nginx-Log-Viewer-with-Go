package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"text/template"
	"time"
)

// LogEntry represents a single log entry
type LogEntry struct {
	IP         string    `json:"ip"`
	UserID     string    `json:"user_id"`
	TimeStamp  time.Time `json:"timestamp"`
	Method     string    `json:"method"`
	RequestURI string    `json:"request_uri"`
	Status     int       `json:"status"`
	UserAgent  string    `json:"user_agent"`
}

func main() {
	// Replace with the actual path to your Nginx log file
	filePath := "sample-small.log"
	// Replace with the desired start and end dates in "2006-01-02 15:04:05" format
	startDateStr := "2024-02-16 11:00:00"
	endDateStr := "2024-02-16 11:59:59"

	startDateTime, err := time.Parse("2006-01-02 15:04:05", startDateStr)
	if err != nil {
		fmt.Println("Error parsing start date:", err)
		return
	}

	endDateTime, err := time.Parse("2006-01-02 15:04:05", endDateStr)
	if err != nil {
		fmt.Println("Error parsing end date:", err)
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Open the log file
		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Create a regular expression to parse the Nginx log format
		logRegex := regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) (\S+) \S+" (\d+) \d+ "([^"]+)" "([^"]+)"`)

		// Track the number of requests per second, RequestURIs, requests per minute, total requests, and RequestURIs per second
		requestsPerSecond := make(map[string]int)
		requestURICounts := make(map[string]int)
		requestsPerMinute := make(map[string]int)
		totalRequests := 0
		requestURIsPerSecond := make(map[string]map[string]int)
		userAgentCounts := make(map[string]int)
		statusCodeCounts := make(map[int]int)
		httpStatusCodes := make(map[int]map[string]int)
		// Create a scanner to read the file line by line
		scanner := bufio.NewScanner(file)

		// Iterate through each line in the log file
		for scanner.Scan() {
			line := scanner.Text()

			// Use the regular expression to extract relevant information
			matches := logRegex.FindStringSubmatch(line)
			if matches != nil {
				timestamp, err := time.Parse("02/Jan/2006:15:04:05 +0700", matches[2])
				if err != nil {
					fmt.Println("Error parsing timestamp:", err)
					continue
				}

				// Check if the entry's date matches the desired date range
				if timestamp.Before(startDateTime) || timestamp.After(endDateTime) {
					continue
				}

				entry := LogEntry{
					IP:         matches[1],
					TimeStamp:  timestamp,
					Method:     matches[3],
					RequestURI: truncateString(matches[4], 100), // Limit RequestURI to 200 characters
					Status:     atoi(matches[5]),
					UserAgent:  matches[7],
				}

				// Count requests per second
				secondKey := entry.TimeStamp.Format("2006-01-02 15:04:05")
				requestsPerSecond[secondKey]++

				// Count RequestURIs
				requestURICounts[entry.RequestURI]++

				// Count requests per minute
				minuteKey := entry.TimeStamp.Format("2006-01-02 15:04")
				requestsPerMinute[minuteKey]++

				// Increment total requests
				totalRequests++

				// Count RequestURIs per second
				if _, ok := requestURIsPerSecond[secondKey]; !ok {
					requestURIsPerSecond[secondKey] = make(map[string]int)
				}
				requestURIsPerSecond[secondKey][entry.RequestURI]++

				// Count User Agents
				userAgentCounts[entry.UserAgent]++

				// Count Status Codes
				statusCodeCounts[entry.Status]++

				// Count http response status codes and corresponding RequestURIs
				if _, ok := httpStatusCodes[entry.Status]; !ok {
					httpStatusCodes[entry.Status] = make(map[string]int)
				}
				httpStatusCodes[entry.Status][entry.RequestURI]++
			}
		}

		// Sort requests per second in descending order
		sortedKeys := make([]string, 0, len(requestsPerSecond))
		for key := range requestsPerSecond {
			sortedKeys = append(sortedKeys, key)
		}
		sort.Slice(sortedKeys, func(i, j int) bool {
			return requestsPerSecond[sortedKeys[i]] > requestsPerSecond[sortedKeys[j]]
		})

		// Sort RequestURIs in descending order
		sortedURIs := make([]string, 0, len(requestURICounts))
		for uri := range requestURICounts {
			sortedURIs = append(sortedURIs, uri)
		}
		sort.Slice(sortedURIs, func(i, j int) bool {
			return requestURICounts[sortedURIs[i]] > requestURICounts[sortedURIs[j]]
		})

		// Sort requests per minute in descending order
		sortedMinutes := make([]string, 0, len(requestsPerMinute))
		for minute := range requestsPerMinute {
			sortedMinutes = append(sortedMinutes, minute)
		}
		sort.Slice(sortedMinutes, func(i, j int) bool {
			return requestsPerMinute[sortedMinutes[i]] > requestsPerMinute[sortedMinutes[j]]
		})

		// Sort User Agents in descending order
		sortedUserAgents := make([]string, 0, len(userAgentCounts))
		for userAgent := range userAgentCounts {
			sortedUserAgents = append(sortedUserAgents, userAgent)
		}
		sort.Slice(sortedUserAgents, func(i, j int) bool {
			return userAgentCounts[sortedUserAgents[i]] > userAgentCounts[sortedUserAgents[j]]
		})

		// Sort Status Codes in descending order
		sortedStatusCodes := make([]int, 0, len(statusCodeCounts))
		for code := range statusCodeCounts {
			sortedStatusCodes = append(sortedStatusCodes, code)
		}
		sort.Slice(sortedStatusCodes, func(i, j int) bool {
			return sortedStatusCodes[i] > sortedStatusCodes[j]
		})

		// Prepare data for the template
		type ViewData struct {
			HttpStatusCodes      map[int]map[string]int
			HttpStatusCodesSlice []struct {
				StatusCode int
				URIs       map[string]int
			}
			Date                      string
			TopRequestsPerSecond      map[string]int
			TopRequestsPerSecondSlice []struct {
				Timestamp string
				Count     int
				URIs      map[string]int
			}
			TopRequestURIs      map[string]int
			TopRequestURIsSlice []struct {
				RequestURI string
				Count      int
			}
			RequestsPerMinute      map[string]int
			RequestsPerMinuteSlice []struct {
				Minute string
				Count  int
			}
			UserAgentCounts      map[string]int
			UserAgentCountsSlice []struct {
				UserAgent string
				Count     int
			}
			TotalRequests          int
			TotalRequestsFormatted string

			StatusCodeCounts      map[int]int
			StatusCodeCountsSlice []struct {
				StatusCode int
				Count      int
			}
		}

		viewData := ViewData{
			Date:                 fmt.Sprintf("%s - %s", startDateTime.Format("2006-01-02 15:04:05"), endDateTime.Format("2006-01-02 15:04:05")),
			TopRequestsPerSecond: requestsPerSecond,
			TopRequestsPerSecondSlice: []struct {
				Timestamp string
				Count     int
				URIs      map[string]int
			}{},
			TopRequestURIs: requestURICounts,
			TopRequestURIsSlice: []struct {
				RequestURI string
				Count      int
			}{},
			RequestsPerMinute: requestsPerMinute,
			RequestsPerMinuteSlice: []struct {
				Minute string
				Count  int
			}{},
			UserAgentCounts: userAgentCounts,
			UserAgentCountsSlice: []struct {
				UserAgent string
				Count     int
			}{},
			TotalRequests:          totalRequests,
			TotalRequestsFormatted: formatNumberWithCommas(totalRequests),

			StatusCodeCounts: statusCodeCounts,
			StatusCodeCountsSlice: []struct {
				StatusCode int
				Count      int
			}{},

			HttpStatusCodes: httpStatusCodes,
			HttpStatusCodesSlice: []struct {
				StatusCode int
				URIs       map[string]int
			}{},
		}

		// Populate the slice for the template (Top Requests Per Second)
		for i, key := range sortedKeys {
			if i >= 10 {
				break
			}
			viewData.TopRequestsPerSecondSlice = append(viewData.TopRequestsPerSecondSlice, struct {
				Timestamp string
				Count     int
				URIs      map[string]int
			}{Timestamp: key, Count: requestsPerSecond[key], URIs: requestURIsPerSecond[key]})
		}

		// Populate the slice for the template (Top RequestURIs)
		for i, uri := range sortedURIs {
			if i >= 10 {
				break
			}
			viewData.TopRequestURIsSlice = append(viewData.TopRequestURIsSlice, struct {
				RequestURI string
				Count      int
			}{RequestURI: uri, Count: requestURICounts[uri]})
		}

		// Populate the slice for the template (Requests Per Minute)
		for i, minute := range sortedMinutes {
			if i >= 10 {
				break
			}
			viewData.RequestsPerMinuteSlice = append(viewData.RequestsPerMinuteSlice, struct {
				Minute string
				Count  int
			}{Minute: minute, Count: requestsPerMinute[minute]})
		}

		// Populate the slice for the template (User Agent Counts)
		for i, userAgent := range sortedUserAgents {
			if i >= 10 {
				break
			}
			viewData.UserAgentCountsSlice = append(viewData.UserAgentCountsSlice, struct {
				UserAgent string
				Count     int
			}{UserAgent: userAgent, Count: userAgentCounts[userAgent]})
		}

		// Populate the slice for the template (HTTP Status Code Counts)
		for i, code := range sortedStatusCodes {
			if i >= 10 {
				break
			}
			viewData.StatusCodeCountsSlice = append(viewData.StatusCodeCountsSlice, struct {
				StatusCode int
				Count      int
			}{StatusCode: code, Count: statusCodeCounts[code]})
		}

		for code, uris := range httpStatusCodes {
			if code != 200 {
				viewData.HttpStatusCodesSlice = append(viewData.HttpStatusCodesSlice, struct {
					StatusCode int
					URIs       map[string]int
				}{StatusCode: code, URIs: uris})
			}

		}

		// Render the template
		tmpl, err := template.New("index").Parse(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Log Analysis Dashboard</title>
			<link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
		</head>
		<body class="bg-gray-100">
		
			<div class="container mx-auto p-4">
		
				<h1 class="text-3xl font-bold text-blue-700 mt-8 mb-4">Log Analysis Dashboard</h1>
		
				<h2 class="text-2xl font-bold text-blue-700 mb-4">Date Range: {{.Date}}</h2>
		
				<h3 class="text-xl font-bold text-blue-700 mb-4">Total Requests</h3>
				<p class="mb-4 font-bold">Total Requests: {{.TotalRequestsFormatted}}</p>
		
				<h3 class="text-xl font-bold text-blue-700 mb-4">Top 10 Requests Per Second for {{.Date}}</h3>
				<table class="border border-collapse border-blue-500 w-full">
					<tr class="bg-blue-200">
						<th class="border border-blue-500 px-4 py-2">Timestamp</th>
						<th class="border border-blue-500 px-4 py-2">Request</th>
						<th class="border border-blue-500 px-4 py-2">RequestURIs (Grouped)</th>
					</tr>
					{{range .TopRequestsPerSecondSlice}}
					<tr>
						<td class="border border-blue-500 px-4 py-2">{{.Timestamp}}</td>
						<td class="border border-blue-500 px-4 py-2">{{.Count}}</td>
						<td class="border border-blue-500 px-4 py-2">
							<table class="border border-collapse border-blue-500 w-full">
								<tr class="bg-blue-100">
									<th class="border border-blue-500 px-4 py-2">RequestURI</th>
									<th class="border border-blue-500 px-4 py-2">Request</th>
								</tr>
								{{range $uri, $count := .URIs}}
								<tr>
									<td class="border border-blue-500 px-4 py-2">{{$uri}}</td>
									<td class="border border-blue-500 px-4 py-2">{{$count}}</td>
								</tr>
								{{end}}
							</table>
						</td>
					</tr>
					{{end}}
				</table>
		
				<h3 class="text-xl font-bold text-blue-700 mt-8 mb-4">Top 10 Request URL for {{.Date}} </h3>
				<table class="border border-collapse border-blue-500 w-full">
					<tr class="bg-blue-200">
						<th class="border border-blue-500 px-4 py-2">RequestURI</th>
						<th class="border border-blue-500 px-4 py-2">Request</th>
					</tr>
					{{range .TopRequestURIsSlice}}
					<tr>
						<td class="border border-blue-500 px-4 py-2">{{.RequestURI}}</td>
						<td class="border border-blue-500 px-4 py-2">{{.Count}}</td>
					</tr>
					{{end}}
				</table>
		
				<h3 class="text-xl font-bold text-blue-700 mt-8 mb-4">Top 10 Requests Per Minute for {{.Date}}</h3>
				<table class="border border-collapse border-blue-500 w-full">
					<tr class="bg-blue-200">
						<th class="border border-blue-500 px-4 py-2">Minute</th>
						<th class="border border-blue-500 px-4 py-2">Request</th>
					</tr>
					{{range .RequestsPerMinuteSlice}}
					<tr>
						<td class="border border-blue-500 px-4 py-2">{{.Minute}}</td>
						<td class="border border-blue-500 px-4 py-2">{{.Count}}</td>
					</tr>
					{{end}}
				</table>
		
				<h3 class="text-xl font-bold text-blue-700 mt-8 mb-4">Toop 10 User Agent Request for {{.Date}}</h3>
				<table class="border border-collapse border-blue-500 w-full">
					<tr class="bg-blue-200">
						<th class="border border-blue-500 px-4 py-2">User Agent</th>
						<th class="border border-blue-500 px-4 py-2">Request</th>
					</tr>
					{{range .UserAgentCountsSlice}}
					<tr>
						<td class="border border-blue-500 px-4 py-2">{{.UserAgent}}</td>
						<td class="border border-blue-500 px-4 py-2">{{.Count}}</td>
					</tr>
					{{end}}
				</table>

				
				<h3 class="text-xl font-bold text-blue-700 mt-8 mb-4">HTTP Status Code Request for {{.Date}}</h1>

				<table class="border border-collapse border-blue-500 w-full">
					<tr class="bg-blue-200">
						<th class="border border-blue-500 px-4 py-2">Status Code</th>
						<th class="border border-blue-500 px-4 py-2">Request</th>
					</tr>
					{{range .StatusCodeCountsSlice}}
					<tr>
						<td class="border border-blue-500 px-4 py-2">{{.StatusCode}}</td>
						<td class="border border-blue-500 px-4 py-2">{{.Count}}</td>
					</tr>
					{{end}}
				</table>

				<h3 class="text-xl font-bold text-blue-700 mt-8 mb-4">Failed HTTP Status Codes for {{.Date}}</h1>

				<table class="border border-collapse border-blue-500 w-full">
					<tr class="bg-blue-200">
						<th class="border border-blue-500 px-4 py-2">Status Code</th>
						<th class="border border-blue-500 px-4 py-2">RequestURIs (Grouped)</th>
					</tr>
					{{range .HttpStatusCodesSlice}}
					<tr>
						<td class="border border-blue-500 px-4 py-2">{{.StatusCode}}</td>
						<td class="border border-blue-500 px-4 py-2">
							<table class="border border-collapse border-blue-500 w-full">
								<tr class="bg-blue-100">
									<th class="border border-blue-500 px-4 py-2">RequestURI</th>
									<th class="border border-blue-500 px-4 py-2">Request</th>
								</tr>
								{{range $uri, $count := .URIs}}
								<tr>
									<td class="border border-blue-500 px-4 py-2">{{$uri}}</td>
									<td class="border border-blue-500 px-4 py-2">{{$count}}</td>
								</tr>
								{{end}}
							</table>
						</td>
					</tr>
					{{end}}
				</table>
		
			</div>
		
		</body>
		</html>
		`)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, viewData)
		if err != nil {
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
	})

	// Start the web server
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// atoi is a helper function to convert a string to an integer
func atoi(s string) int {
	i := 0
	for _, c := range s {
		i = i*10 + int(c-'0')
	}
	return i
}

// truncateString truncates the input string to the specified length
func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

func formatNumberWithCommas(n int) string {
	str := strconv.Itoa(n)
	length := len(str)

	if length <= 3 {
		return str
	}

	// Calculate the number of commas needed
	numCommas := (length - 1) / 3

	// Create a buffer to hold the formatted string
	result := make([]byte, length+numCommas)

	// Copy digits to buffer and insert commas
	for i, j := 0, 0; i < length; i++ {
		result[j] = str[i]
		j++
		if (length-i-1)%3 == 0 && i != length-1 {
			result[j] = '.'
			j++
		}
	}

	return string(result)
}
