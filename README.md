##Info Hash##

Make sure you don't have a newline at the end of the file, you may also want to make sure it ends with an 'e'.
The info-hash of a torrent file is the SHA-1 hash of the info-section (in bencoded form) from the .torrent file. Essentially you need to decode the file (it's bencoded) and remember the byte offsets where the content of the value associated with the "info" key begins and end. That's the range of bytes you need to hash.
For example, if this is the torrent file:
d4:infod6:pieces20:....................4:name4:test12:piece lengthi1024ee8:announce27:http://tracker.com/announcee
You wan to just hash this section:
d6:pieces20:....................4:name4:test12:piece lengthi1024ee

##Check Vailid UTF-8 for strings##

```
func main() {
    validUTF8 := []byte("Hello, 世界")  // Valid UTF-8
    invalidUTF8 := []byte{0xff, 0xfe, 0xfd} // Invalid UTF-8

    fmt.Println("Valid UTF-8:", utf8.Valid(validUTF8))   // Output: true
    fmt.Println("Invalid UTF-8:", utf8.Valid(invalidUTF8)) // Output: false
}
```