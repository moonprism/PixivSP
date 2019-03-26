package main

import "github.com/moonprism/PixivSP/lib"

func main() {
	p := NewPixiv(lib.PixivConf.PixivUser, lib.PixivConf.PixivPasswd)
	p.Login()
}
