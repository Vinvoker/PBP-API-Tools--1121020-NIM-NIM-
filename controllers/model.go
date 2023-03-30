package controllers

type Order struct {
	IDorder        int `json:"id"`
	WaktuTransaksi int `json:"waktu_transaksi"`
}

type OrderDetails struct {
	IDorder  int `json:"id_order"`
	IDproduk int `json:"id_produk"`
	Quantity int `json:"quantity"`
}

type Produk struct {
	IDproduk   int    `json:"id_produk"`
	NamaProduk string `json:"nama_produk"`
	Harga      int    `json:"harga"`
	Gambar     string `json:"gambar"`
}

type DataOwners struct {
	IDowner    int    `json:"id_owner"`
	NamaOwner  string `json:"nama_owner"`
	EmailOwner string `json:"email"`
}

type Receiver struct {
	OwnerName  string `json:"name"`
	OwnerEmail string `json:"email"`
}

type Message struct {
	ProductName string
	Quantity    int
	Price       int
}
