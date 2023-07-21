package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

const (
	responseOpeningToken   = "{\"tags\":["
	responseChunkSeparator = ","
	responseClosingToken   = "]}"

	openingTableXMLTagRegex = "<table*"
	closingTableXMLTagRegex = `.*</table>`

	internalServerError = "internal server error"
)

func main() {
	http.HandleFunc("/tags", getJSONTags)

	fmt.Println("Serving on port 8080...")
	panic(http.ListenAndServe(":8080", nil))
}

func getJSONTags(w http.ResponseWriter, r *http.Request) {
	respCtrl := http.NewResponseController(w)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")

	cmd := exec.CommandContext(r.Context(), "exiftool", "-listx")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("get stdout pipe from command failed: %s\n", err)
		writeErrorResponse(w, http.StatusInternalServerError, internalServerError)
	}

	chunk := make(chan string)
	done := make(chan bool)
	go func() {
		scanner := bufio.NewScanner(stdout)
		var buffer strings.Builder

		openingTableTag := regexp.MustCompile(openingTableXMLTagRegex)
		closingTableTag := regexp.MustCompile(closingTableXMLTagRegex)

		bufferData := false
		for scanner.Scan() {
			if openingTableTag.MatchString(scanner.Text()) {
				bufferData = true
			}

			if bufferData {
				buffer.WriteString(scanner.Text())
			}

			if closingTableTag.MatchString(scanner.Text()) {
				chunk <- buffer.String()
				buffer.Reset()
				bufferData = false
			}
		}
		done <- true
	}()

	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	defer func() {
		err := cmd.Wait()
		if err != nil {
			log.Printf("wait for command completion failed: %s\n", err)
		}
	}()

	firstTable := true
	for {
		select {
		case <-r.Context().Done():
			log.Printf("Connection interrupted by client. Closing command stdout pipe...\n")
			err := stdout.Close()
			if err != nil {
				log.Printf("close stdout pipe from command failed: %s\n", err)
			}
			return

		case c := <-chunk:
			tag, err := xmlTableDataToJSONTag(c)
			if err != nil {
				log.Printf("conversion from xml table data to JSON tag failed: %s\n", err)
				writeErrorResponse(w, http.StatusInternalServerError, internalServerError)
			}

			err = writeJSONTag(w, respCtrl, tag, firstTable)
			if err != nil {
				log.Printf("write JSON tag failed: %s\n", err)
				writeErrorResponse(w, http.StatusInternalServerError, internalServerError)
			}

		case <-done:
			_, err = w.Write([]byte(responseClosingToken))
			if err != nil {
				log.Printf("write response closing token failed: %s\n", err)
				writeErrorResponse(w, http.StatusInternalServerError, internalServerError)
			}

			err = respCtrl.Flush()
			if err != nil {
				log.Printf("flushing buffer failed: %s\n", err)
				writeErrorResponse(w, http.StatusInternalServerError, internalServerError)
			}
			return
		}
	}
}

func writeErrorResponse(w http.ResponseWriter, statusError int, errorMsg string) {
	msg := struct {
		Error string `json:"error"`
	}{
		Error: errorMsg,
	}

	enc := json.NewEncoder(w)

	w.WriteHeader(statusError)
	enc.Encode(msg)
}

func writeJSONTag(w http.ResponseWriter, c *http.ResponseController, tag JSONTag, firstTable bool) error {
	if firstTable {
		_, err := w.Write([]byte(responseOpeningToken))
		if err != nil {
			return fmt.Errorf("failed to print response opening token: %w", err)
		}
	} else {
		_, err := w.Write([]byte(responseChunkSeparator))
		if err != nil {
			return fmt.Errorf("failed to print response chunk separator: %w", err)
		}
	}

	b, err := json.Marshal(tag)
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	if err != nil {
		return err
	}

	err = c.Flush()
	if err != nil {
		return err
	}

	return nil
}
