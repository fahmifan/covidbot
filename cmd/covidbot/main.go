package main

func main() {
	rootCMD.AddCommand(crawlerCMD(), parseCMD(), testerCMD())
	rootCMD.Execute()
}
