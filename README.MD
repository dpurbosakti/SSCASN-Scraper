This project only for personal uses and education purpose

# How to clone this repository

```bash
git clone https://github.com/dpurbosakti/SSCASN-Scraper.git
cd SSCASN-Scraper
go mod tidy
```

# Before running this project
Set kodeRefPend 
Set namaJurusan

# How to get KodeRefPend

1. Buka browser dan masuk ke halaman https://sscasn.bkn.go.id/
2. Tekan f12 dan klik search bar lalu masukkan nama jurusan yang diinginkan (contoh: DIII Fisioterapi)
3. Buka tab source dan cari field "cepat_kode"

# How to run this project

```bash
go run main.go
```

# How to use this project

## Flags

- `-provinsi`: Provinsi yang diinginkan. Contoh: `-provinsi="Jawa Timur"`

## How to run this project with flags

```bash
go run main.go -provinsi="Jawa Timur"
```