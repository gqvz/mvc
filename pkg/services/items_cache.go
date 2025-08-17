package services

var itemsCache string

func GetItemsCache() string {
	return itemsCache
}

func SetItemsCache(jsonString string) {
	itemsCache = jsonString
}

func ClearItemsCache() {
	itemsCache = ""
}
