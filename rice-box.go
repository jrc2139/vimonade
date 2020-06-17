package main

import (
	"time"

	"github.com/GeertJohan/go.rice/embedded"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    "package.go",
		FileModTime: time.Unix(1592394757, 0),

		Content: string("// package certs\n\n/* go:generate go run github.com/phogolabs/parcello/cmd/parcello -r */\n"),
	}
	file3 := &embedded.EmbeddedFile{
		Filename:    "service.key",
		FileModTime: time.Unix(1592395069, 0),

		Content: string("-----BEGIN RSA PRIVATE KEY-----\nMIIJKAIBAAKCAgEA3BAFYKs6pFNfBCE+G/jwfphUCZ4RDGaFTtkYvmy7TmuwUoQj\nky4Gxkw9MtKIuRGX+IImRwBABvgML7f0c41BlXzKcILT7XdFUrtdTfNOHXLXnj0A\nt+7l2kDvVjglHGrXhtmkj2TXtWaum7UdbafGOSNr8GcM3ngd3ql38FoXOYNi8+TF\n34QbSjHDztt0YT3Vl6rxAqWFArBdbichEzrOLrChrezp2ILVosd89MyFf9Q/B5Ka\nvrtbFJsrj+FyTdMPfByJCXjVqoZGM61AHzuYQbnpBqUjgcV6S3ThM09deY8uCdEP\nnzaMwan7v7BgOgx4LM6oMbrErTy9jSE/qe2mknhYOWyafMDGKeM4n3k6pqsKly8V\nqcnCyzXawR83MtPzXOzNp/XaOI8jCDbKs/dWIMXqAvXF/DLMHgddclSCPAZw4hHx\nMyUX8gGELMq8H8jRA2ncCjvGWI7/4E2xkz7irj87qwCSW/IhCaJvXQMaPRD2pC+H\nDh01PvyKouyGMfWXhjdCX4z3gIqe7SdtBA4KAjagYMsbxlMaXOI7QFJCuntAItk5\nlCBCN1I2YYmU/xTiZhrhBpgT0wYJsKtHII4BoncbRMzb13vHy8myvL8zEcGmAWVt\nas0gheAhYGJMOoGL5d88/35i3JXNcw8WETMONECaiqHGf4EfAWKtdYyAU9UCAwEA\nAQKCAgBTeBgyFICHt7/Ad0NxpLjDq8+kXjolM/W4dIv4HpHlKt+UTN6Zgsm7CUvl\nLQoC+HNmJSeTorqmagtlRicIdTm9E7pNdrIfzi+TM9IuMO0eKmMyd/z/xzOT1zFw\nyJb68mORmJfcc+mFus7b7JSe9RYwsgjNBKlS8MiTnkAKAIHypK0xyUJsh1U6jSmy\nGndIMwYDIymLUXDkFjU6BMl827884m5Q5B3Gq8mQlo4E5FZ2p/DIVIkxcysW48xg\nJKkQb8/dyY0I9LZGmeuiykIaFqant3Z1uMmA/YBngouwxJe19eMGgi+kujIleu4s\nRSAapIQoqrINOsRC2VkkYMofEC8vhuWmfEbPSjQEqye47Q7C0ZXvSu3CrYBEnxHV\nwfZwfGtGPJqJe1GXKn5QaK+dXhSIqcDIwcXfa1OkbkLk2CWL6tMOM95d27V+OM7W\nG5ya2xMwaUJ38Fp/PiSOuVV4TlzwfYeLLfHrlqqbhaSdviZro0iSompzNy2k+YRH\nEgvNB25ostBYTeIvP6usV2ZpXCVyzrlxTf5FXHrCLVQDPiqQZNrRhOfT6xQR4rk1\n9icwhK8l+m0x0GSpSWYWu7+O2frKrguxKqypUXVk+ZRgd54fadPXTpo+A4CAzDIv\nKtWpn23hp/iKpof+o+Z/DJdobXljTEJ0E3SyReNdC1FbeBJNeQKCAQEA/EvfaJty\nz2B6uZ8FzgnaXZ5IRGKB1kgHiqGQyqR3lLHqYqVFBa2b1Vs8lFFowC4ykoNX8Jq/\nWPpBw16YGboDKY8FwiDfFTf9A8I+FnK1BvyzX2uLFVCxcuEjW1fGybqrWdxT263n\nANA0Yye7U5W6LTp40opnAucQ1bZG0JpHnsQAVVzVKjJCpJHm5U5lB2RlI0+1LJNS\nAOiQqi18SGA1kUsxTeF3+p6URaHlrSuRUJoc+oNN1TDQsdX/XcWl/ndTUSFmbfxI\nGweEDqRdwQp20/71lHgHOKgbogHqjdTM5JFQ5qm6/uxF5Nn/hXkj3CkYYbqW7hu/\nlfySQQSDnc7TPwKCAQEA30sDl2VEI75N2ai3tjFQcvv2loeY9CuMCw52TOi4IMuQ\naBUWElquxI7bJ4z75ksPsEQExrEV0+sITY852clvyop6QrVO1GB2W5ttXwIF5aBf\nUbIrC1Qv03n+kd9abzYnFH98KYPxsUxci9ZgA7donGKIOZHMomF/qaDMqxUOnk25\nPRwm+338MhsnY0XkTvmo+o4R2RL0h5SoEgJZYrN4Sz6YMolj2z41GB+mkKozC8co\nAsbYplBXWkjS4KXOQEQg4P4ud/EMB5mYNrTgbKKCFsBOlk2O8HOChPIiAqLsn6QV\nf2v2TPkunZ6KbiHEZ3TzRw6UgOMwL8m73m1e1aZX6wKCAQAtwdz87eR+s/LOI4c7\n/RF7lS9qJ6uAn4OuourNtdJyR2pJBcxk4T24DloIVFN5N2e4ptWWL5qwmoK+2jMf\nx1q3eNcEhE2xXXwn6Fy2WYt3fvFRRwHslbv5J9fvwxWslIxrOciDuSCCR0CZEyWo\nXSls9oPfO3a/UgT9nZduUezXYJjm4nVOt9raWhPUVsl/87dcFiK3uOhQfd1u390A\ni2JrvYVtqIzICWa+0kQDijlKswi6boH5Pmc7OaKc8THP2vhjaHlZTT4OmOhcd3cB\ngdJXVJBZowM8RVDtqwdNPeEDO0++5d2iSlvKy7bKEFRuo41mfB7PhHzUyQAFhroQ\nLuilAoIBAHaSTfDp/FoCpzJqrktYOoEknRfoH2ehbDc+0cEbXxNDJYavk83hS6bi\nuStyaR0sRMN0Cxk7Vfz3dKxC3xRwLCXgjPW5c4fBRXh1u4lU+K6sD5HBS6wzY0Yo\nJO9vLIWbuvrei588Cm78vrQe/VNb5HgOtonji0e7AGCiG6zJfL7BRRlXRrgLeY1d\n7/d+WLM7Tejm4kFkGGean/kYOED6TmmebpF/dYAps2YBAKEXUA30DqIS117RkOFH\nhHt4cGKeCtuO/jwAy0OJ41NBj18AmJXePpz/yGSU4f0Y2siNnZtUNXo5aUwMkh1u\n39GFqtbJOppD+sXKXn8x38pIR7CqKUsCggEBANt/2vDO0Ynr1+mCCmeHUIljfqWU\nm/0tUpEIt8j2JRxtNzhv9aukdbGFiK/056aGXB7lR5enu6XiV+0WAd6PvxNzKhYB\n5Pv1yK1mNOhaU+vPFB92C4F/QYxsYguA0XaznrF4j3+pY8xnDhlZxj1y/7njVcG4\nRWlWMe0BEsvamd/rb1wpTbQhM52rXKY360Nea0WFLFw7+NhyfF/jMHUpvIZrxU4g\nKH+eczhdqwzdEqxLESD4e8uwd9qyM8o6nrsAxH5RlfEbUY7HmsGjAFk84OGy6agJ\n0GBD80JRHH77+wq3UGqfYWkD+Q6xB0o1JJSKgZ+7Re7V14PbPau3IEN4KIY=\n-----END RSA PRIVATE KEY-----\n"),
	}
	file4 := &embedded.EmbeddedFile{
		Filename:    "service.pem",
		FileModTime: time.Unix(1592395069, 0),

		Content: string("-----BEGIN CERTIFICATE-----\nMIIFIzCCAwugAwIBAgIJANUcI71F4apzMA0GCSqGSIb3DQEBCwUAMC0xCzAJBgNV\nBAYTAlVTMQswCQYDVQQIDAJOSjERMA8GA1UECgwIQ0EsIEluYy4wHhcNMjAwNjE3\nMTE1NzQ5WhcNMjEwNjE3MTE1NzQ5WjBDMQswCQYDVQQGEwJVUzELMAkGA1UECAwC\nTkoxEzARBgNVBAoMClRlc3QsIEluYy4xEjAQBgNVBAMMCWxvY2FsaG9zdDCCAiIw\nDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBANwQBWCrOqRTXwQhPhv48H6YVAme\nEQxmhU7ZGL5su05rsFKEI5MuBsZMPTLSiLkRl/iCJkcAQAb4DC+39HONQZV8ynCC\n0+13RVK7XU3zTh1y1549ALfu5dpA71Y4JRxq14bZpI9k17Vmrpu1HW2nxjkja/Bn\nDN54Hd6pd/BaFzmDYvPkxd+EG0oxw87bdGE91Zeq8QKlhQKwXW4nIRM6zi6woa3s\n6diC1aLHfPTMhX/UPweSmr67WxSbK4/hck3TD3wciQl41aqGRjOtQB87mEG56Qal\nI4HFekt04TNPXXmPLgnRD582jMGp+7+wYDoMeCzOqDG6xK08vY0hP6ntppJ4WDls\nmnzAxinjOJ95OqarCpcvFanJwss12sEfNzLT81zszaf12jiPIwg2yrP3ViDF6gL1\nxfwyzB4HXXJUgjwGcOIR8TMlF/IBhCzKvB/I0QNp3Ao7xliO/+BNsZM+4q4/O6sA\nklvyIQmib10DGj0Q9qQvhw4dNT78iqLshjH1l4Y3Ql+M94CKnu0nbQQOCgI2oGDL\nG8ZTGlziO0BSQrp7QCLZOZQgQjdSNmGJlP8U4mYa4QaYE9MGCbCrRyCOAaJ3G0TM\n29d7x8vJsry/MxHBpgFlbWrNIIXgIWBiTDqBi+XfPP9+YtyVzXMPFhEzDjRAmoqh\nxn+BHwFirXWMgFPVAgMBAAGjMDAuMCwGA1UdEQQlMCOCCWxvY2FsaG9zdIcQAAAA\nAAAAAAAAAAAAAAAAAYcEfwAAATANBgkqhkiG9w0BAQsFAAOCAgEAE4QdruQC3gEP\njQLp43ctby+IPwRa7OBKJVHg6dYEQyyAZ37JELBcLH8TM+llTnmga871UO4XMwBr\nV3RlfSdq0owyTvce+bjHHB+p4avOkTVHzwHbQZUW2xbrp0YvM47phudU+iarkMCf\nvMMBvCQu9QG/rva/H1EEuG0qJ1yAJTArg6kj86FbfTWpGuACnJzBKXYjpgvS+u5+\n2ZQZ317XF6RqUrJDA1orCeTUD2tYYZ5s4sxd/kpRmjHDncVwTi2Ia4Oc0AhpRsfe\nIWAkvAHbCg0cuqJEwjdK9ctQeWGdped9uFaUPNcxpdIMGe7veYP0+ttpIfXKfdzX\n1O4xFAUIeqN7hOcuj6cjk7u9ioWebtXRTOXUJz8aj0f1XJris8Wp3suh5eGZIHxP\nh2X1VhPCiyAwcipnjq9hbyJeXCZl+EhmnoyOY+v89IA3RLnvhmyEkV+BbZ11l6bo\n2YQZFZNhNr1Yhf2uILPbAMILP1KNO4KSxc6PFy1J/jN9ITqdS00E8eDyeyhJ2MAx\nxvP5E8ZZEnke/WOOXsj4iyBdhbobsk1XA9DZjQjLMiZOWZCfSX2iNnnkjV428zm2\n3eH2EB5XHXmL506zCoxPfZJCJMMQBmP8K6vNWn2m2YVZpXwFGpZSTb8PNSoesPLJ\nVLQjGHj1Z4qTlMMFdcC1Uhn9pPYYq/s=\n-----END CERTIFICATE-----\n"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1592394768, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // "package.go"
			file3, // "service.key"
			file4, // "service.pem"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`certs`, &embedded.EmbeddedBox{
		Name: `certs`,
		Time: time.Unix(1592394768, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"package.go":  file2,
			"service.key": file3,
			"service.pem": file4,
		},
	})
}
