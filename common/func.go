package common

import (
	"bytes"
	"crypto/md5"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	httpurl "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GetLocalIP get local network ip
func GetLocalIP() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0, len(addrs))
	for _, address := range addrs {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}

	return ips, nil
}

// AuthCodeType auth code type
type AuthCodeType int

const (
	// ENCODE encode str
	ENCODE AuthCodeType = iota
	// DECODE decode str
	DECODE
)

// Method http request method
type Method int8

const (
	// POST post http request
	POST Method = iota
	// GET http request get method
	GET
)

func (m Method) String() string {
	var name string
	switch m {
	case POST:
		name = "POST"
	case GET:
		name = "GET"
	default:
		name = "UNKNOWN"
	}

	return name
}

// AbsolutePath get execute binary path
func AbsolutePath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}

	return filepath.Dir(path), nil
}

// Md5Sum md5
func Md5Sum(text string) string {
	h := md5.New()
	io.WriteString(h, text)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// NewRand return *rand.Rand
func NewRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// RandRangeInt return min<=x<max
func RandRangeInt(min, max int) int {
	if max == min {
		return min
	}
	if max < min {
		min, max = max, min
	}
	return min + NewRand().Intn(max-min)
}

// RandRangeInt32 return min<=x<max
func RandRangeInt32(min, max int32) int32 {
	if max == min {
		return min
	}
	if max < min {
		min, max = max, min
	}
	return min + NewRand().Int31n(max-min)
}

// Reverse string reverse
func Reverse(s string) string {
	b := []byte(s)
	n := ""
	for i := len(b); i > 0; i-- {
		n += string(b[i-1])
	}
	return string(n)
}

// RandArray rand string slice
func RandArray(arr []string) string {
	return arr[NewRand().Intn(len(arr))]
}

// RsaEncode rsa encode
func RsaEncode(b, rsaKey []byte) ([]byte, error) {
	block, _ := pem.Decode(rsaKey)
	if block == nil {
		return b, errors.New("key error")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return b, err
	}
	return rsa.EncryptPKCS1v15(crand.Reader, pub.(*rsa.PublicKey), b)
}

// RsaDecode rsa decode
func RsaDecode(b, rsaKey []byte) ([]byte, error) {
	block, _ := pem.Decode(rsaKey)
	if block == nil {
		return b, errors.New("key error")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return b, err
	}
	return rsa.DecryptPKCS1v15(crand.Reader, priv, b)
}

// IsIP ip address is valid
func IsIP(ip string) bool {
	ips := strings.Split(ip, ".")
	if len(ips) != 4 {
		return false
	}
	for _, v := range ips {
		i, err := strconv.Atoi(v)
		if err != nil {
			return false
		}
		if i < 0 || i > 255 {
			return false
		}
	}

	return true
}

// IsMac mac address is valid
func IsMac(mac string) bool {
	if len(mac) != 17 {
		return false
	}
	mac = strings.ToLower(mac)

	r := `^(?i:[0-9a-f]{1})(?i:[02468ace]{1}):(?i:[0-9a-f]{2}):(?i:[0-9a-f]{2}):(?i:[0-9a-f]{2}):(?i:[0-9a-f]{2}):(?i:[0-9a-f]{2})`
	reg, err := regexp.Compile(r)
	if err != nil {
		return false
	}
	m := reg.FindStringSubmatch(mac)
	if m == nil {
		return false
	}

	return true
}

// Base64Encode string encode
func Base64Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// Base64Decode string decode
func Base64Decode(str string) ([]byte, error) {
	x := len(str) * 3 % 4
	switch {
	case x == 2:
		str += "=="
	case x == 1:
		str += "="
	}
	return base64.StdEncoding.DecodeString(str)
}

// RangeArray generate array
func RangeArray(m, n int) (b []int) {
	if m >= n || m < 0 {
		return b
	}

	c := make([]int, 0, n-m)
	for i := m; i < n; i++ {
		c = append(c, i)
	}

	return c
}

// Authcode Discuz Authcode golang version
// params[0] encrypt/decrypt bool true：encrypt false：decrypt, default: false
// params[1] key
// params[2] expires time(second)
// params[3] dynamic key length
func Authcode(text string, params ...interface{}) (str string, err error) {
	l := len(params)

	isEncode := DECODE
	key := "abcdefghijklmnopqrstuvwxyz1234567890"
	expiry := 0
	cKeyLen := 8

	if l > 0 {
		isEncode = params[0].(AuthCodeType)
	}

	if l > 1 {
		key = params[1].(string)
	}

	if l > 2 {
		expiry = params[2].(int)
		if expiry < 0 {
			expiry = 0
		}
	}

	if l > 3 {
		cKeyLen = params[3].(int)
		if cKeyLen < 0 {
			cKeyLen = 0
		}
	}
	if cKeyLen > 32 {
		cKeyLen = 32
	}

	timestamp := time.Now().Unix()

	// md5sum key
	mKey := Md5Sum(key)

	// keyA encrypt
	keyA := Md5Sum(mKey[0:16])
	// keyB validate
	keyB := Md5Sum(mKey[16:])
	// keyC dynamic key
	var keyC string
	if cKeyLen > 0 {
		if isEncode == ENCODE {
			// encrypt generate a key
			keyC = Md5Sum(fmt.Sprint(timestamp))[32-cKeyLen:]
		} else {
			// decrypt get key from header of string
			keyC = text[0:cKeyLen]
		}
	}

	// generate encrypt/decrypt key
	cryptKey := keyA + Md5Sum(keyA+keyC)
	// key length
	keyLen := len(cryptKey)
	if isEncode == ENCODE {
		// The first 10 strings is expires time
		// 10-26 strings is validator strings
		var d int64
		if expiry > 0 {
			d = timestamp + int64(expiry)
		}
		text = fmt.Sprintf("%010d%s%s", d, Md5Sum(text + keyB)[0:16], text)
	} else {
		// get strings except dynamic key
		b, e := Base64Decode(text[cKeyLen:])
		if e != nil {
			return "", e
		}
		text = string(b)
	}

	// text length
	textLen := len(text)
	if textLen <= 0 {
		err = fmt.Errorf("auth [%s] textLen <= 0", text)
		return
	}

	// keys
	box := RangeArray(0, 256)
	//
	rndKey := make([]int, 0, 256)
	cryptKeyB := []byte(cryptKey)
	for i := 0; i < 256; i++ {
		pos := i % keyLen
		rndKey = append(rndKey, int(cryptKeyB[pos]))
	}

	j := 0
	for i := 0; i < 256; i++ {
		j = (j + box[i] + rndKey[i]) % 256
		box[i], box[j] = box[j], box[i]
	}

	textB := []byte(text)
	a := 0
	j = 0
	result := make([]byte, 0, textLen)
	for i := 0; i < textLen; i++ {
		a = (a + 1) % 256
		j = (j + box[a]) % 256
		box[a], box[j] = box[j], box[a]
		result = append(result, byte(int(textB[i])^(box[(box[a]+box[j])%256])))
	}

	if isEncode == ENCODE {
		// trim equal
		return keyC + Base64Encode(result), nil
	}

	// check expire time
	d, e := strconv.ParseInt(string(result[0:10]), 10, 0)
	if e != nil {
		err = fmt.Errorf("expires time error: %s", e.Error())
		return
	}

	if (d == 0 || d-timestamp > 0) && string(result[10:26]) == Md5Sum(string(result[26:]) + keyB)[0:16] {
		return string(result[26:]), nil
	}

	err = fmt.Errorf("Authcode text [%s] error", text)
	return
}

// TimeFormat format time.Time
func TimeFormat(t time.Time, f int) (timeStr string) {
	switch f {
	case 0:
		timeStr = t.Format("2006-01-02 15:04:05")
	case 1:
		timeStr = t.Format("2006-01-02")
	case 2:
		timeStr = t.Format("15:04:05")
	case 3:
		timeStr = t.Format("20060102150405")
	case 4:
		timeStr = t.Format("2006-01-02 15:04")
	}

	return
}

// Now format now
func Now(f ...int) string {
	var format int
	if len(f) > 0 {
		format = f[0]
	} else {
		format = 0
	}
	return TimeFormat(time.Now(), format)
}

// StructToMap struct convert to map
func StructToMap(item interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	if item == nil {
		return res
	}
	v := reflect.TypeOf(item)
	reflectValue := reflect.ValueOf(item)
	reflectValue = reflect.Indirect(reflectValue)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		tag := v.Field(i).Tag.Get("json")
		field := reflectValue.Field(i).Interface()
		if tag != "" && tag != "-" {
			tag = strings.Split(tag, ",")[0]
			if v.Field(i).Type.Kind() == reflect.Struct {
				res[tag] = StructToMap(field)
			} else {
				res[tag] = field
			}
		}
	}
	return res
}

// Keys key of map
func Keys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

// Values value of map
func Values(m map[string]interface{}) []interface{} {
	values := make([]interface{}, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}

	return values
}

// IsEmpty true: nil, "", false, 0, 0.0, {}, []
func IsEmpty(val interface{}) (b bool) {
	if val == nil {
		return true
	}
	v := reflect.ValueOf(val)

	switch v.Kind() {
	case reflect.Bool:
		b = (val.(bool) == false)
	case reflect.String:
		b = (len(val.(string)) == 0)
	case reflect.Array, reflect.Slice, reflect.Map:
		b = (v.Len() == 0)
	default:
		b = (v.Interface() == reflect.ValueOf(0).Interface() || v.Interface() == reflect.ValueOf(0.0).Interface())
	}

	return b
}

// IP2Long IP convert to long int
func IP2Long(ipstr string) (uint64, error) {
	var ip uint64
	r := `^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})`
	reg, err := regexp.Compile(r)
	if err != nil {
		return 0, err
	}
	ips := reg.FindStringSubmatch(ipstr)
	if ips == nil {
		return 0, fmt.Errorf("Error ip addr:" + ipstr)
	}

	ipInt := make([]int, 0, 4)
	for index, i := range ips {
		d, err := strconv.Atoi(i)
		if err != nil {
			return 0, nil
		}
		if d < 0 || d > 255 {
			return 0, fmt.Errorf("Error ip addr:%s in segment[%d]", ipstr, index)
		}
		ipInt = append(ipInt, d)
	}

	ip += uint64(ipInt[0] * 0x1000000)
	ip += uint64(ipInt[1] * 0x10000)
	ip += uint64(ipInt[2] * 0x100)
	ip += uint64(ipInt[3])

	return ip, nil
}

// Long2IP longint convert to IP
func Long2IP(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", ip>>24, ip<<8>>24, ip<<16>>24, ip<<24>>24)
}

// Contains item is in map/slice/array
func Contains(haystack interface{}, needle interface{}) bool {
	targetValue := reflect.ValueOf(haystack)
	switch reflect.TypeOf(haystack).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == needle {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(needle)).IsValid() {
			return true
		}
	}

	return false
}

// Trim remove "", \r, \t, \n
func Trim(str string) string {
	return strings.Trim(str, " \r\n\t")
}

// Split split by match
func Split(str, match string) []string {
	re := regexp.MustCompile(match)
	return re.Split(str, -1)
}

// SplitBySpaceTab splite by space or tab
func SplitBySpaceTab(str string) []string {
	return Split(str, `[ \t]+`)
}

// HTTPRequest request
// url request url
// method request method post or get
// args[0] type is map[string]string or string, request paramaters, \x00@ if upload file
// args[1] type is map[string]string, request headers
// args[2] type is bool, whether to return the result
// args[3] type is *http.Client, custom client
func HTTPRequest(url string, method Method, args ...interface{}) (string, error) {
	paramsMap := make(map[string]string) // request parameters
	var paramsStr string
	var paramsIsStr bool
	headers := make(map[string]string) // request headers
	rtn := true
	var client *http.Client
	var ok bool
	argsLen := len(args)
	if argsLen > 0 {
		paramsMap, ok = args[0].(map[string]string)
		if !ok {
			paramsStr, ok = args[0].(string)
			if !ok {
				return "", errors.New("Params error")
			}
			paramsIsStr = true
		}
	}
	if argsLen > 1 {
		headers, ok = args[1].(map[string]string)
		if !ok {
			return "", errors.New("Headers error")
		}
	}
	if argsLen > 2 {
		rtn, ok = args[2].(bool)
		if !ok {
			return "", errors.New("Return bool error")
		}
	}
	if argsLen > 3 {
		client, ok = args[3].(*http.Client)
		if !ok {
			return "", errors.New("Http client error")
		}
	} else {
		client = http.DefaultClient
	}

	var req *http.Request
	var err error
	contentType := "" // default content-type
	var queryString string
	if paramsIsStr {
		queryString = URLEncode(paramsStr)
	} else {
		queryString = URLEncode(paramsMap)
	}
	if method == GET {
		// GET
		if queryString != "" {
			if strings.Index(url, "?") != -1 {
				// has params
				url += "&" + queryString
			} else {
				// no params
				url += "?" + queryString
			}
		}

		req, err = http.NewRequest("GET", url, nil)
	} else {
		// POST
		// whether there is upload file
		var isFile bool
		if !paramsIsStr {
			for _, v := range paramsMap {
				if strings.Index(v, "\x00@") == 0 {
					// there is upload file
					isFile = true
					break
				}
			}
		}
		if isFile {
			bodyBuf := new(bytes.Buffer)
			bodyWriter := multipart.NewWriter(bodyBuf)

			for key, value := range paramsMap {
				if strings.Index(value, "\x00@") == 0 {
					value = strings.Replace(value, "\x00@", "", -1)
					fileWriter, err := bodyWriter.CreateFormFile(key, filepath.Base(value))
					if err != nil {
						return "", err
					}
					fh, err := os.Open(value)
					if err != nil {
						return "", err
					}
					defer fh.Close()

					// iocopy
					_, err = io.Copy(fileWriter, fh)
					if err != nil {
						return "", err
					}
				} else {
					bodyWriter.WriteField(key, value)
				}
			}

			// Important if you do not close the multipart writer you will not have a terminating boundry
			bodyWriter.Close()
			contentType = bodyWriter.FormDataContentType()
			req, err = http.NewRequest("POST", url, bodyBuf)
		} else {
			if paramsIsStr {
				contentType = "application/json; charset-utf-8"
				req, err = http.NewRequest("POST", url, strings.NewReader(paramsStr))
			} else {
				contentType = "application/x-www-form-urlencoded; charset=utf-8"
				req, err = http.NewRequest("POST", url, strings.NewReader(queryString))
			}

		}
	}

	if err != nil {
		return "", err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// add headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if rtn {
		// need return
		bData, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", err
		}

		return string(bData), nil
	}

	return "", nil
}

// CopyFile cp command
func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return 0, err
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return 0, err
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

// URLEncode urlencode
func URLEncode(params interface{}) string {
	q, ok := params.(string)
	if ok {
		return httpurl.QueryEscape(q)
	}
	m, ok := params.(map[string]string)
	if ok {
		val := httpurl.Values{}
		for k, v := range m {
			val.Set(k, v)
		}

		return val.Encode()
	}

	return ""
}

// URLDecode urldecode
func URLDecode(str string) (string, error) {
	return httpurl.QueryUnescape(str)
}

// UUID uuid
func UUID() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	return u.String(), nil
}

// IsMobile check chinese mobile number
func IsMobile(mobile string) bool {
	reg := `^1([34578][0-9])\d{8}$`
	rgx := regexp.MustCompile(reg)

	return rgx.MatchString(mobile)
}

// Exists whether file or directory exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// IsDir return bool
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFile return tool
func IsFile(path string) bool {
	return !IsDir(path)
}

// SHA256 return string
func SHA256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}
