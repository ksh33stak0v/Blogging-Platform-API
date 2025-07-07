package main

func main() {
	db := DBConnect()
	CreateTable(db)
	defer db.Close()
}