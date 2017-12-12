// Copyright © 2017 GRAWLER rahulrumalla@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/rahulrumalla/grawler/crawler"
	"github.com/spf13/cobra"
)

var (
	workerCount         int
	includeStaticAssets bool
)

// crawlCmd represents the crawl command
var crawlCmd = &cobra.Command{
	Use:   "crawl",
	Short: "Crawls the url and generates a sitemap",
	Long: `Crawls the url and generates a sitemap
	
Usage:
======

grawler https://monzo.com -i -w 4`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Must supply exactly 1 url to crawl")
		}

		u, err := url.ParseRequestURI(args[0])
		if err != nil {
			log.Fatalf("Error parsing uri from command-line argument - %s", args[0])
			return err
		}

		grawler := crawler.NewGrawler(u, workerCount, includeStaticAssets)
		start := time.Now()
		siteNodes := grawler.Crawl()

		elapsed := time.Since(start)

		counter := 0
		for k, v := range siteNodes {
			if v.IsCrawled {
				fmt.Printf("%s [Crawled status: %v]\n", k, v.IsCrawled)
				counter++

				for a := range v.InternalLinks {
					fmt.Printf("  └── [link] %s\n", a)
				}

				if includeStaticAssets {
					for a := range v.InternalStaticAssets {
						fmt.Printf("  └── [static asset] %s\n", a)
					}
				}
			}
		}

		fmt.Printf("Found and crawled %d links in  %s\n", counter, elapsed)

		return nil
	},
}

func init() {
	RootCmd.AddCommand(crawlCmd)

	crawlCmd.Flags().IntVarP(&workerCount, "workerCount", "w", 4, "sets the number of workers crawling")
	crawlCmd.Flags().BoolVarP(&includeStaticAssets, "includeStaticAssets", "i", false, "sets the flag to include or exclude the static assets while crawling")
}
