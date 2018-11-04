# Connect your TP-Link outlets to Homekit

## Building and running

Check out repository and run `go build`. Run the binary. Please specify the outlets' hostnames using the `TPLINK_OUTLET_HOSTS` env variable (as a comma separated string). It supports multiple devices, but remember to give them distinct hostnames in your network (or alternatively specify IP addresses in the variable).

## Known compatible devices

* [HS100](https://www.tp-link.com/us/products/details/cat-5516_HS100.html)

If you have tested more, please send a pull request.

## Credits

This has been made possible by the [great work of softScheck](https://github.com/softScheck/tplink-smartplug) who reverse engineered the protocol and the "encryption".
