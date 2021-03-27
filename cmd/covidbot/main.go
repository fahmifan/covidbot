package main

func main() {
	rootCMD.AddCommand(crawlerCMD(), parseCMD())
	rootCMD.Execute()
}
