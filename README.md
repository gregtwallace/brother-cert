# Brother Cert
Brother Cert is a command line tool to automatically install an ssl 
certificate on a brother printer.

## Compatibility Notice

The tool was build specifically for the Brother MFC-L2710DW, and MFC-L2750DW printers. It 
seems reasonable other Brother printers probably use the same mechanisms but your mileage may vary.

## Usage

The tool will connect to the printer, convert the pem files into p12 format,
upload the p12 to the printer, activate the new p12 and https, restart 
the printer, and then finally delete the previously active cert.

Run the tool as:

`./brother-cert --hostname printer.example.com --password secret --keyfile key.pem --certfile cert.pem [FLAGS]`

Help can be viewed with:

`./brother-cert --help`

### Initial SSL Setup

This will configure your Brother printer to use SSL so Cert Warden can manage the certificate from now on out.

1. In Cert Warden create an RSA and NOT ECDSA private key.
1. Let the certificate generate and use this to download it 1curl -H "X-API-Key: CERTAPIKEY.PRIVATEKEYAPIKEY" https://certwarden.example.com/certwarden/api/v1/download/pfx/printer --output brother.pem1
1. Go to Security> Certificate and upload the Cert and upload the brother.pem, the password PRIVATEKEYAPIKEY
Under Security> Ca Certificate extract the CA via `openssl pkcs12 -info -in brother.pem -nokeys` and upload it
1. Under Security> TLS Settings set the required TLS Settings. I suggest a minimum of TLS 1.2.
1. Under Network> Protocol> "HTTP Server Settings" select the correct SSL cert and reboot the printer. Now it will recognize the new SSL cert.

## Note About Install Automation

The application supports passing all args instead as environmenT variables by prefixing the flag name with `BROTHER_CERT`. 

e.g. `BROTHER_CERT_KEYPEM`

### Required ARGs

- `BROTHER_CERT_KEYPEM={{PRIVATE_KEY_PEM}}`
- `BROTHER_CERT_CERTPEM={{CERTIFICATE_PEM}}`
- `BROTHER_CERT_HOSTNAME=printer.example.com`
- `BROTHER_CERT_PASSWORD=secret`

There are mutually exclusive flags that allow specifying the pem 
as either filenames or directly as strings. The strings are useful 
for passing the pem content from another application without having 
to save the pem files to disk.

Putting all of this together, you can combine the install binary with 
a tool like Cert Warden (https://www.certwarden.com/) to call the 
install binary, with environment variables, to directly upload new 
certificates as they're issued by Cert Warden, without having to write a 
separate script.

![Cert Warden with Brother Cert](https://raw.githubusercontent.com/gregtwallace/brother-cert/master/img/brother-cert.png)
