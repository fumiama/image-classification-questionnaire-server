<div align="center">
  <img src=".github/ichiko.jpg" width = "360" height = "360" alt="Ichiko"><br>
  <h1>图像分类问卷调查服务器(ICQS)</h1>
  基于go/python的图像分类调查服务器<br><br>
</div>

> 用户可以利用api自己上传图像并指定分类，也可以从服务器的图像中查看图片，并指定分类。

<div align=center> <a href="#"> <img src="https://counter.seku.su/cmoe?name=icqs&theme=gb" /> </a> </div>

# 使用准备

## Python准备

你需要安装`pillow`，`numba`，`imagehash`，`quart/flask`, `gevent`(如果使用`flask`)等包以确保程序运行。

```bash
pip install -r requirements.txt
```

## Golang准备
由于整合了[setu-class](https://github.com/fumiama/setu-class)，你需要编译安装[setu-class-cpp](https://github.com/fumiama/setu-class-cpp)库，该库要求您已经安装了`libtorch`。然后，请将该仓库的`ero.pt`和`nor.pt`复制到本项目编译出的可执行文件旁。

另外，您还需要部署一个[simple-storage](https://github.com/fumiama/simple-storage)，命令行参数中需要填写其`apiurl`与`authkey`。

接下来安装`libwebp-dev`，并在使用前运行
```bash
go mod tidy
```

# 开始使用

首先克隆本仓库
```bash
git clone --depth=1 https://github.com/fumiama/image-classification-questionnaire-server.git
```
### Golang版
#### 1. 命令行参数
1. 如果你是`ubuntu`用户，由于该系统绑定`80`端口需要`root`权限，因此需要添加可选参数`userid`以在绑定端口后降权运行。
2. 密码必须为两个汉字，在运行后密码将在命令行被隐藏，但不会清除命令历史记录，请手动清除。
```bash
Usage: <listen_addr> <apiurl> <password> <authkey> <dbfile> (userid) &
```
注意：
1. 如果添加末尾的`&`，程序将会以`daemon`运行。
2. `userid`为可选项。如果设置，程序将会在绑定端口后切换到该`uid`处理请求。
3. `Windows`下使用不支持`userid`选项。
4. `dbfile`位置任意。
#### 2. 编译
- 如果使用`gc`，添加`-ldflags "-s -w"`编译即可。
- 如果使用`gccgo`，推荐使用如下优化参数以达到最佳效果。
```bash
go build -compiler gccgo -gccgoflags "-Wa,--strip-local-absolute,-R -s -Wl,-x,-X,--sort-common,--enable-new-dtags,--hash-style=gnu -fno-bounds-check -freg-struct-return -O3" -o im
```

### Python版
1. 如果你是`ubuntu`用户，由于该系统绑定`80`端口需要`root`权限，因此需要添加可选参数`server_uid`以在绑定端口后降权运行。
2. 密码文件`pwd_path`必须为以`UTF16BE`编码存储的两个汉字（包括文件头`0xfeff`），总长`6`字节。

#### 简易版`server.py`的语法如下

**注意**：使用一段时间后可能会无响应，目前尚未解决，只能通过`daemon.sh`监控进程是否退出，如果退出则重新拉起。

```bash
./server.py [-d] <user_dir> <image_dir> <pwd_path> (server_uid)
```

其中：
1. `-d`为可选项，如果设置，程序将会以`daemon`运行。
2. `server_uid`为可选项。如果设置，程序将会在绑定端口后切换到该`uid`处理请求。

#### `Quart/Flask`版`server_quart/flask.py`的语法如下（高并发）

```bash
./server_quart/flask.py <user_dir> <image_dir> <pwd_path> (server_uid) 2>&1 > ./log.txt &
```

其中：
1. 如果添加末尾的`&`，程序将会以`daemon`运行。
2. `server_uid`为可选项。如果设置，程序将会在绑定端口后切换到该`uid`处理请求。

注意:

1. 服务端图片扩展名只接受`.webp`，客户端上传时任意。如需其它格式请自行修改代码。
2. 图片的唯一标识使用了该图片`dhash`值的`base16384`编码的前五个汉字。

# Golang版API（推荐）

对应执行文件为`server.go`，该版本的class只能为0~7的整数。

### 0. 直接访问

- 格式: http://[server_domain]/

- 说明: 直接通过简易网页访问服务。

### 1. 注册用户

- 格式: http://[server_domain]/signup?key=1234567890

- 返回:
1. 成功
```json
{"stat":"success", "id":"%XX%XX%XX%XX%XX%XX"}
```
2. 密码错误
```json
{"stat":"wrong", "id":"null"}
```
3. 处理错误
```json
400 BAD REQUEST
```
- 说明:

1. 返回转义的两个`utf-8`编码的汉字，代表下面用到的`uuid`
2. `key`后跟10位整数，表示密码与当前秒数异或的结果，与服务端相差10秒内有效

### 2. 指定分类(投票)

- 格式: http://[server_domain]/vote?uuid=用户&img=投票的图片&class=n

- 返回:
1. 成功
```json
{"stat":"success"}
```
2. 图片名格式非法
```json
{"stat":"invimg"}
```
3. 用户名格式非法
```json
{"stat":"invid"}
```
4. 找不到uuid/img/class字段
```json
400 BAD REQUEST
```
5. class非法
```json
{"stat":"invclass"}
```

- 说明:

1. `uuid`字段容纳两(数量不可增减)个`utf-8`编码的汉字，表示投票用户。
2. `img`字段容纳五个(数量不可增减)`utf-8`编码的汉字，唯一标识了这张图片。
3. `class`字段class只能为0~7的整数，代表该图片所属标签。

### 3. 下载图片

- 格式: http://[server_domain]/img?path=某一张图片

- 返回:
1. 成功
```
图片数据
```
2. 无此图片
```json
{"stat":"nosuchimg"}
```
3. 图片名格式非法
```json
{"stat":"invimg"}
```

- 说明: `目标的图片`是五个(数量不可增减)`utf-8`编码的汉字，唯一标识了这张图片。

### 4. 上传图片

- 格式: `HTTP POST`到http://[server_domain]/upload?uuid=用户

- 返回: 
1. 成功
```json
{"stat":"success","result":[{"stat": "success","img": "%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX"}]}
```
2. 相似或相同图片存在
```json
{"stat":"success","result":[{"stat":"exist","img":"%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX"}]}
```
3. 空请求体
```json
{"stat":"emptybody"}
```
4. 找不到此用户
```json
{"stat":"noid"}
```
5. 解析dhash错误
```json
{"stat":"success","result":[{"stat":"notanimg","img":""}]}
```
6. 不是图片
```json
{"stat":"success","result":[{"stat":"notanimg","img":""}]}
```
7. io错误
```json
{"stat":"success","result":[{"stat":"ioerr","img":""}]}
```
8. 图片转码错误
```json
{"stat":"success","result":[{"stat":"encerr","img":""}]}
```

- 说明: 必须为`webp`、`jpg`、`png`或`gif`格式。使用`wget`时，可使用如下命令。

```bash
wget --post-file=image.webp http://[server_domain]/upload?uuid=用户
```

### 5. 以表单形式上传图片

> 该API无法上传合计大于64M的文件，可以更改ParseMultipartForm入参以提高上限

- 格式: `HTTP POST`到http://[server_domain]/upform?uuid=用户

该格式支持批量上传。

1. 成功
返回时，会将全部结果统一送回。
```json
{"stat":"success","result":[{"name":"a.jpg","stat":"exist"},{"name":"b.jpg","stat":"exist"},{"name":"c.jpg","stat":"exist"},{"name":"d.jpg","stat":"success"},{"name":"e.jpg","stat":"success"}]}
```
其中，result列表的每一项都遵循条目4的result格式。
2. 找不到此用户
```json
{"stat":"noid"}
```
3. io错误
```json
{"stat":"success","result":[{"stat":"ioerr","img":""}]}
```
- 说明: 必须为`webp`、`jpg`、`png`或`gif`格式。

### 6. 从未投票图片中随机选择图片并返回图片数据
- 格式: http://[server_domain]/pickdl?uuid=用户
- 返回:
1. 成功
```
图片数据
```
2. 无更多图片
```json
{"stat":"nomoreimg"}
```
3. 无此用户
```json
{"stat":"noid"}
```
4. 用户名格式非法
```json
400 BAD REQUEST
```
- 说明: `用户`是两个(数量不可增减)`utf-8`编码的汉字，唯一标识了某个用户。

### 7. 从未投票图片中随机选择图片并返回图片名

- 格式: http://[server_domain]/pick?uuid=用户
- 返回:
1. 成功
```json
{"stat":"success", "img":"%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX", "uploader":"%XX%XX%XX%XX%XX%XX"}
```
2. 无更多图片
```json
{"stat":"nomoreimg"}
```
3. 无此用户
```json
{"stat":"noid"}
```
4. 用户名格式非法
```json
400 BAD REQUEST
```

- 说明:
1. `用户`是两个(数量不可增减)`utf-8`编码的汉字，唯一标识了某个用户。
2. 返回的图片名经过了转义。

### 8. dice
参数详见[setu-class](https://github.com/SakuraACGN/setu-class)

### 9. 获取图片 dhash
- 格式: http://[server_domain]/dhash?pidp=pid_pn

- 返回:
1. 成功
```json
{"PidP":"53538084_p0","UID":63652,"Width":650,"Height":906,"Title":"秋月照月本表紙","Author":"中乃空","R18":false,"Tags":"[\"艦隊これくしょん\",\"舰队collection\",\"C89\",\"秋月\",\"Akizuki\",\"照月\",\"Teruzuki\",\"艦ぱい\",\"shipgirl breasts\",\"尻神様\",\"尻神样\",\"即夜戦\",\"即将夜战\",\"秋月型\",\"Akizuki-class\",\"ねじ込みたい尻\",\"这屁股让人想肛\"]","Ext":"png","DatePath":"2015/11/14/00/28/41","DHash":"嗉聚裌蠼嬀","Md5":"34f4ed2a6500dc8c6822f0d4333639ba"}
```
2. 失败
```
各种状态码，附带简短说明
```

# 简易版API（不建议用）

对应执行文件为`server.py`

### 0. 直接访问

- 格式: http://[server_domain]/index.html

- 说明: 直接通过简易网页访问服务。

### 1. 注册用户

- 格式: http://[server_domain]/signup?1234567890

- 返回: 成功(succ)，密码错误(null)，处理错误(erro)

- 说明:

1. 返回`utf-8`编码的两个汉字，代表下面用到的`uuid`
2. 后跟10位整数，表示密码与当前秒数异或的结果，与服务端相差10秒内有效

### 2. 指定分类(投票)

- 格式: http://[server_domain]/vote?uuid=用户&img=投票的图片&class=n

- 返回: 成功(succ)，错误(erro)

- 说明:

1. `uuid`字段容纳两(数量不可增减)个`utf-8`编码的汉字，表示投票用户。
2. `img`字段容纳五个(数量不可增减)`utf-8`编码的汉字，唯一标识了这张图片。
3. `class`字段容纳最多184个`ASCII`编码，代表该图片所属标签。

### 3. 下载图片

- 格式: http://[server_domain]/目标的图片

- 返回: 成功(图片数据)，空(null)，错误(erro)

- 说明: `目标的图片`是五个(数量不可增减)`utf-8`编码的汉字，唯一标识了这张图片。

### 4. 上传图片

- 格式: `HTTP POST`到http://[server_domain]/upload?uuid=用户

- 返回: 成功(succ)，错误(erro)，图片相似/无此用户(null)

- 说明: 必须为`webp`、`jpg`、`png`或`gif`格式。使用`wget`时，可使用如下命令。

```bash
wget --post-file=image.webp http://[server_domain]/upload?uuid=用户
```

### 5. 从未投票图片中随机选择图片并返回图片数据

- 格式: http://[server_domain]/pickdl?用户

- 返回: 成功(图片数据)，空(null)，错误(erro)

- 说明: `用户`是两个(数量不可增减)`utf-8`编码的汉字，唯一标识了某个用户。

### 6. 从未投票图片中随机选择图片并返回图片名

- 格式: http://[server_domain]/pick?用户

- 返回: 成功(图片名)，空(null)，错误(erro)

- 说明:
1. `用户`是两个(数量不可增减)`utf-8`编码的汉字，唯一标识了某个用户。
2. 返回的图片名经过了转义。

# Quart/Flask版API

对应执行文件为`server_quart.py`

### 0. 直接访问

- 格式: http://[server_domain]/index.html

- 说明: 直接通过简易网页访问服务。

### 1. 注册用户

- 格式: http://[server_domain]/signup?key=1234567890

- 返回:
1. 成功
```json
{"stat":"success", "id":"%XX%XX%XX%XX%XX%XX"}
```
2. 密码错误
```json
{"stat":"wrong", "id":"null"}
```
3. 处理错误
```json
{"stat":"error", "id":"null"}
```
- 说明:

1. 返回转义的两个`utf-8`编码的汉字，代表下面用到的`uuid`
2. `key`后跟10位整数，表示密码与当前秒数异或的结果，与服务端相差10秒内有效

### 2. 指定分类(投票)

- 格式: http://[server_domain]/vote?uuid=用户&img=投票的图片&class=n

- 返回:
1. 成功
```json
{"stat":"success"}
```
2. 图片名格式非法
```json
{"stat":"invimg"}
```
3. 用户名格式非法
```json
{"stat":"invid"}
```
4. 找不到此用户
```json
{"stat":"noid"}
```

- 说明:

1. `uuid`字段容纳两(数量不可增减)个`utf-8`编码的汉字，表示投票用户。
2. `img`字段容纳五个(数量不可增减)`utf-8`编码的汉字，唯一标识了这张图片。
3. `class`字段容纳最多184个`ASCII`编码，代表该图片所属标签。

### 3. 下载图片

- 格式: http://[server_domain]/img?path=某一张图片

- 返回:
1. 成功
```
图片数据
```
2. 读图片错误
```json
{"stat":"readimgerr"}
```
3. 无此图片
```json
{"stat":"nosuchimg"}
```
4. 图片名格式非法
```json
{"stat":"invimg"}
```

- 说明: `目标的图片`是五个(数量不可增减)`utf-8`编码的汉字，唯一标识了这张图片。

### 4. 上传图片

> Quart版目前无法上传大于20M的图片

- 格式: `HTTP POST`到http://[server_domain]/upload?uuid=用户

- 返回: 
1. 成功
```json
{"stat":"success", "img":"%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX"}
```
2. 相似或相同图片存在
```json
{"stat":"exist", "img":"%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX"}
```
3. 用户名格式非法
```json
{"stat":"invid"}
```
4. 找不到此用户
```json
{"stat":"noid"}
```
5. 接收图片错误
```json
{"stat":"recverr"}
```
6. 不是图片
```json
{"stat": "notanimg"}
```

- 说明: 必须为`webp`、`jpg`、`png`或`gif`格式。使用`wget`时，可使用如下命令。

```bash
wget --post-file=image.webp http://[server_domain]/upload?uuid=用户
```

### 5. 以表单形式上传图片

> 该API目前无法上传合计大于20M的文件

- 格式: `HTTP POST`到http://[server_domain]/upform?uuid=用户

该格式支持批量上传。

- 返回: 

返回时，会将全部结果统一送回。

```json
{"result":[{"name":"a.jpg","stat":"exist"},{"name":"b.jpg","stat":"exist"},{"name":"c.jpg","stat":"exist"},{"name":"d.jpg","stat":"success"},{"name":"e.jpg","stat":"success"}]}
```

其中，列表的每一项都遵循如下格式

1. 成功
```json
{"name":"xxx","stat":"success"}
```
2. 相似或相同图片存在
```json
{"name":"xxx","stat":"exist"}
```
3. 用户名格式非法
```json
{"name":"xxx","stat":"invid"}
```
4. 找不到此用户
```json
{"name":"xxx","stat":"noid"}
```
5. 接收图片错误
```json
{"stat":"recverr"}
```
6. 不是图片
```json
{"stat": "notanimg"}
```

- 说明: 必须为`webp`、`jpg`、`png`或`gif`格式。

### 6. 从未投票图片中随机选择图片并返回图片数据

- 格式: http://[server_domain]/pickdl?uuid=用户

- 返回:
1. 成功
```
图片数据
```
2. 读图片错误
```json
{"stat":"readimgerr"}
```
3. 无更多图片
```json
{"stat":"nomoreimg"}
```
4. 无此用户
```json
{"stat":"noid"}
```
5. 用户名格式非法
```json
{"stat":"invid"}
```

- 说明: `用户`是两个(数量不可增减)`utf-8`编码的汉字，唯一标识了某个用户。

### 7. 从未投票图片中随机选择图片并返回图片名

- 格式: http://[server_domain]/pick?uuid=用户

- 返回:
1. 成功
```json
{"stat":"success", "img":"%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX%XX", "uploader":"%XX%XX%XX%XX%XX%XX"}
```
2. 读图片错误
```json
{"stat":"readimgerr"}
```
3. 无更多图片
```json
{"stat":"nomoreimg"}
```
4. 无此用户
```json
{"stat":"noid"}
```
5. 用户名格式非法
```json
{"stat":"invid"}
```

- 说明:
1. `用户`是两个(数量不可增减)`utf-8`编码的汉字，唯一标识了某个用户。
2. 返回的图片名经过了转义。

# 小工具

一些实用的小工具放在了`tools`文件夹，包括批量上传图片，批量转换图片到`webp`，批量重命名文件，批量缩小`webp`大小，从flask/quart迁移到go等。

使用方法详见注释。
