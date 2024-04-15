# Brother Cert
Brother Cert is a command line tool to automatically install an ssl 
certificate on a brother printer.

## Compatibility Notice

The tool was build specifically for the Brother MFC-L2710DW printer. It 
seems reasonable other Brother printers probably use the same mechanisms
but your mileage may vary.

## Usage

The tool will connect to the printer, conver the pem files into p12 format,
upload the p12 to the printer, activate the new p12 and https, restart 
the printer, and then finally delete the previously active cert.

Run the tool as:

`./brother-cert --hostname printer.example.com --password secret --keyfile key.pem --certfile cert.pem [FLAGS]`

Help can be viewed with:

`./brother-cert --help`

## Note About Install Automation

The application supports passing all args instead as environment 
variables by prefixing the flag name with `BROTHER_CERT`. 

e.g. `BROTHER_CERT_KEYPEM`

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
