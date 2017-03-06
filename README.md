lookup
======

Go library to load data into structs. Steps:

* Define tag `"lookup"` in struct fields. The value should consist of the key to lookup.
* Optional trailing `",optional"` means the lookup won't fail if the key is missing.
* Provide an extraction function, e.g, `os.LookupEnv`. Use `NoError`/`NoBool` to adapt functions with different signatures.
* Lookup sequences may be defined. Typically the last step contains the defaults using `Map`.

Encoding:

* complex64 and complex128: separate r,i with comma.
* []byte: base64.

---
Author: Carlos Eduardo Lenz  
License: MIT
