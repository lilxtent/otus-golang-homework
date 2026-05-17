package hw10programoptimization

import (
	"archive/zip"
	"testing"
)

var benchmarkGetDomainStatResult DomainStat

func BenchmarkGetDomainStat(b *testing.B) {
	usersArchive, err := zip.OpenReader("testdata/users.dat.zip")
	if err != nil {
		b.Fatal(err)
	}
	defer usersArchive.Close()

	usersFile := usersArchive.File[0]

	b.ReportAllocs()
	b.SetBytes(int64(usersFile.UncompressedSize64))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, err := usersFile.Open()
		if err != nil {
			b.Fatal(err)
		}

		stat, err := GetDomainStat(data, "biz")
		closeErr := data.Close()
		if err != nil {
			b.Fatal(err)
		}
		if closeErr != nil {
			b.Fatal(closeErr)
		}

		benchmarkGetDomainStatResult = stat
	}
}
