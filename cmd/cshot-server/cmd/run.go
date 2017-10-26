// Copyright © 2017 apokalyptik
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"log"

	"github.com/apokalyptik/cshot/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the cshot service",
	Run: func(cmd *cobra.Command, args []string) {
		var srv = &service.Server{
			Host:   viper.GetString("host"),
			Port:   viper.GetInt("port"),
			Chrome: viper.GetString("chrome"),
		}
		log.Fatal(srv.ListenAndServe(viper.GetInt("procs")))
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().String("host", "0.0.0.0", "address to listen on")
	viper.BindPFlag("host", runCmd.PersistentFlags().Lookup("host"))
	runCmd.PersistentFlags().Int("port", 80, "port to listen on")
	viper.BindPFlag("port", runCmd.PersistentFlags().Lookup("port"))
	runCmd.PersistentFlags().Int("procs", 10, "chrome processes to run")
	viper.BindPFlag("procs", runCmd.PersistentFlags().Lookup("procs"))
}
