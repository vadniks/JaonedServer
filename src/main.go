
package main

import "JaonedServer/network"

func main() {
    xNetwork := network.Init()
    xNetwork.ProcessClients()
}
