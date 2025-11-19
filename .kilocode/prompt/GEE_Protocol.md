# GEE Protocol

1. 先 POST 请求 `https://kh.google.com/geauth?ct=pro`, 请求体为 AuthBodyHexString(Hex 值字符串需要转化为字节)

2. 接收第一步的响应(字节二进制响应), 并按照以下方法提取出 `SessionID`:

```js
const arrayBuffer = (window._arrayBuffer = await res_body.clone().arrayBuffer())
let str_arrayBuffer = ""
// console.log("长度", arrayBuffer.byteLength);
switch (arrayBuffer.byteLength) {
  case 112:
    str_arrayBuffer = String.fromCharCode.apply(
      null,
      new Uint8Array(arrayBuffer.slice(8, 88)) //截取长度是未经验证的
    )
    break
  case 124:
    str_arrayBuffer = String.fromCharCode.apply(
      null,
      new Uint8Array(arrayBuffer.slice(8, 100))
    )
    break
  case 136:
    str_arrayBuffer = String.fromCharCode.apply(
      null,
      new Uint8Array(arrayBuffer.slice(8, 112))
    )
    break
  case 144:
    str_arrayBuffer = String.fromCharCode.apply(
      null,
      new Uint8Array(arrayBuffer.slice(8, 120))
    )
    break
  default:
    debugger
}
// console.log(
//   "完整字符串",
//   String.fromCharCode.apply(null, new Uint8Array(arrayBuffer))
// );
// console.log("截出来的cookie", str_arrayBuffer);
document.cookie = `SessionId="${str_arrayBuffer}"; path=/;`
// console.log("document.cookie ", document.cookie);
```

3. 启动一个 Goroutine 每隔 2 分钟进行上面的 POST 请求, 以保持 SessionID 的有效性

4. 写一个函数, 用于转发 GEE 请求, 并在其中添加 cookie 信息: `SessionId=${SessionID}; path=/;`

```go
func GEERelay(path string, methods string, body io.Reader) (*http.Response, error) {

}
```
