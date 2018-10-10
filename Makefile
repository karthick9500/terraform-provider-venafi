TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

all: build test testacc

#TODO: build for other platfroms
build: clean fmtcheck
	go test ./...
	go build
	terraform init

clean:
	rm -fv terraform.tfstate*
	rm -fv terraform-provider-venafi

test: fmtcheck
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4
testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

#Integration tests using real terrafomr binary
test_e2e: test_e2e_fake test_e2e_tpp test_e2e_cloud

test_e2e_tpp:
	echo yes|terraform apply -target=venafi_certificate.tpp_certificate
	terraform state show venafi_certificate.tpp_certificate
	terraform output cert_certificate_tpp > /tmp/cert_certificate_tpp.pem
	cat /tmp/cert_certificate_tpp.pem
	cat /tmp/cert_certificate_tpp.pem|openssl x509 -inform pem -noout -issuer -serial -subject -dates

test_e2e_cloud:
	echo yes|terraform apply -target=venafi_certificate.cloud_certificate
	terraform state show venafi_certificate.cloud_certificate
	terraform output cert_certificate_cloud > /tmp/cert_certificate_cloud.pem
	cat /tmp/cert_certificate_cloud.pem
	cat /tmp/cert_certificate_cloud.pem|openssl x509 -inform pem -noout -issuer -serial -subject -dates

test_e2e_fake:
	echo yes|terraform apply -target=venafi_certificate.fake_certificate
	terraform state show venafi_certificate.fake_certificate
	terraform output cert_certificate_fake > /tmp/cert_certificate_fake.pem
	cat /tmp/cert_certificate_fake.pem
	cat /tmp/cert_certificate_fake.pem|openssl x509 -inform pem -noout -issuer -serial -subject -dates
	terraform output cert_private_key_fake > /tmp/cert_private_key_fake.pem
	cat /tmp/cert_private_key_fake.pem

test_e2e_fake_ecdsa:
	echo yes|terraform apply -target=venafi_certificate.fake_ecdsa_certificate
	terraform state show venafi_certificate.fake_ecdsa_certificate
	terraform output cert_certificate_fake_ecdsa > /tmp/cert_certificate_fake_ecdsa.pem
	cat /tmp/cert_certificate_fake_ecdsa.pem
	cat /tmp/cert_certificate_fake_ecdsa.pem|openssl x509 -inform pem -noout -issuer -serial -subject -dates
	terraform output cert_private_key_fake_ecdsa > /tmp/cert_private_key_fake_ecdsa.pem
	cat /tmp/cert_private_key_fake_ecdsa.pem