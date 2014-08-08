[RFC2616](https://www.ietf.org/rfc/rfc2616.txt)

Augmented BNF:

CHAR            = <any US-ASCII character (octets 0 - 127)>
OCTET           = <any 8-bit sequence of data>

CR              = <US-ASCII CR, carriage return (13)>
LF              = <US-ASCII LF, linefeed (10)>
CRLF            = CR LF
SP              = <US-ASCII SP, space (32)>
HT              = <US-ASCII HT, horizontal-tab (9)>
LWS             = [CRLF] 1*( SP | HT )
CTL             = <any US-ASCII control character (octets 0 - 31) and DEL (127)>
TEXT            = <any OCTET except CTLs, but including LWS>

token           = 1*<any CHAR except CTLs or separators>
quoted-pair     = "\" CHAR
ctext           = <any TEXT excluding "(" and ")">
comment         = "(" *( ctext | quoted-pair | comment ) ")"

product-version = token
product         = token ["/" product-version]
User-Agent      = "User-Agent" ":" 1*( product | comment )
