package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"
)

func newAddCommand() *cobra.Command {

	var address string
	var port uint
	var key, value string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an event",
		Long:  `Add an event to the authenticated data structure`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Adding key [ %s ] with value [ %s ]\n", key, value)

			url := fmt.Sprintf("http://%s:%d/events", address, port)
			jsonStr := map[string]string{"key": key, "value": value}
			jsonValue, _ := json.Marshal(jsonStr)
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
			if err != nil {
				return err
			}

			defer resp.Body.Close()
			fmt.Printf("Reponse Status: %s\n", resp.Status)
			fmt.Printf("Reponse Headers: %v\n", resp.Header)
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Printf("Reponse Body: %v\n", string(body))
			return nil
		},
	}

	cmd.Flags().StringVarP(&address, "address", "a", "", "Server's HTTP API host")
	cmd.Flags().UintVarP(&port, "port", "p", 0, "Server's HTTP API port")
	cmd.Flags().StringVarP(&key, "key", "k", "", "Key to add")
	cmd.Flags().StringVarP(&value, "value", "v", "", "Value to add")
	cmd.MarkFlagRequired("address")
	cmd.MarkFlagRequired("port")
	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("value")

	return cmd
}
