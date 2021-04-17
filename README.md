# 图像分类问卷调查服务器

一个基于python的图像分类调查服务器，用户可以利用api自己上传图像并指定分类，也可以从服务器的图像中查看图片，并指定分类。

# 使用准备

由于引入了其他仓库，因此需要使用`--recursive`进行克隆

```bash
git clone --recursive https://github.com/fumiama/image-classification-questionnaire-server.git
```

接下来，你需要安装`cmake`，然后执行以下命令以生成程序所需的C库。

```bash
mkdir build
cd build
cmake ..
make
```

# API

注意:

1. 图片扩展名只接受`.webp`，如需其它格式请自行修改代码。
2. 图片的唯一标识使用了该图片`md5`值的`base16384`编码的前四个汉字。

### 0. 直接访问

格式: http://[server_domain]/index.html

说明: 直接通过简易网页访问服务。

### 1. 注册用户

格式: http://[server_domain]/signup

说明: 返回`utf-8`编码的两个汉字，代表下面用到的`uuid`

### 2. 指定分类(投票)

格式: http://[server_domain]/vote?uuid=用户&img=投票图片&class=n

返回: 成功(succ)，错误(erro)

说明:

1. `uuid`字段容纳两(数量不可增减)个`utf-8`编码的汉字，表示投票用户。
2. `img`字段容纳四个(数量不可增减)`utf-8`编码的汉字，唯一标识了这张图片。
3. `class`字段容纳任意数量的`ASCII`编码，代表该图片所属标签。

### 3. 下载图片

格式: http://[server_domain]/目标图片

返回: 成功(图片数据)，空(null)，错误(erro)

说明: `目标图片`是四个(数量不可增减)`utf-8`编码的汉字，唯一标识了这张图片。

### 4. 上传图片

格式: `HTTP POST`到http://[server_domain]/upload

返回: 成功(succ)，错误(erro)

说明: 图片必须为`webp`格式。使用`wget`时，可使用如下命令。

```bash
wget --post-file=image.webp http://[server_domain]/upload
```

### 5. 从未投票图片中随机选择图片并返回图片数据

格式: http://[server_domain]/pickdl?用户

返回: 成功(图片数据)，空(null)，错误(erro)

说明: `用户`是两个(数量不可增减)`utf-8`编码的汉字，唯一标识了某个用户。

### 6. 从未投票图片中随机选择图片并返回图片名

格式: http://[server_domain]/pick?用户

返回: 成功(图片名)，空(null)，错误(erro)

说明:
1. `用户`是两个(数量不可增减)`utf-8`编码的汉字，唯一标识了某个用户。
2. 返回的图片名经过了转义。