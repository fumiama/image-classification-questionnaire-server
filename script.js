function AddUser(name, value, day) {
    var d = new Date();
    d.setTime(d.getTime()+(day*24*60*60*1000));
    var 止 = "expires="+d.toGMTString();
    document.cookie = name + "=" + value + "; " + 止;
}
function GetUser(name) {
    var name = name + "=";
    var ca = document.cookie.split(';');
    for(var i=0; i<ca.length; i++) {
        var c = ca[i].trim();
        if (c.indexOf(name)==0) return c.substring(name.length,c.length);
    }
    return "";
}
me = GetUser("uuid");
function GetMsg(url, doing) {
    var request = new XMLHttpRequest();	//第一步：建立所需的对象
    request.open('GET', url, true);		//第二步：打开连接,将request参数写在网址中
    request.send();						//第三步：发送request
    request.onreadystatechange = function () {
        if (request.readyState == 4 && request.status == 200) {
            doing(request.responseText);
        }
    };
}
function GetImg(){
    if(me == "") alert("未登录!");
    else {
        GetMsg("/pick?uuid=" + me, function rri(t) {
            j = JSON.parse(t)
            if(j.stat == "success") document.getElementById("img_display").src = "/img?path=" + j.img;
            else if(j.stat == "nomoreimg") alert("无更多图片!");
            else alert("随机失败，请重试");
        });
    }
}
function LogIn() {
    username = prompt("请输入用户名，错误的用户名无法加载图片","示例");
    if(username != null) {
        if(username.length == 2) {
            me = username;
            AddUser("uuid", me, 7);
        }
        else if(username.length == 0) document.cookie = me = "";
    }
}
function Encode(text) {
    text = escape(text.toString()).replace(/\+/g, "%2B");
    var re = text.match(/(%([0-9A-F]{2}))/gi);
    if (re) {
        for (var bit = 0; bit < re.length; bit++) {
            var code = re[bit].substring(1,3);
            if (parseInt(code, 16) >= 128) {
                text = text.replace(re[bit], '%u00' + code);
            }
        }
    }
    text = text.replace('%25', '%u0025');
    return text;
}
function HexToDec(hex_num) {
    			var length = hex_num.length, string = new Array(length), code;
    			for (var bit = 0; bit < length; bit++) {
        			code = hex_num.charCodeAt(bit);
        			if (48<=code && code < 58) code -= 48;
        			else code = (code & 0xdf) - 65 + 10;
        			string[bit] = code;
    			}
    			return string.reduce(function(sum, remain) {
        			sum = 16 * sum + remain;
        			return sum;
    			}, 0);
}
function Reg() {
    if(me == "") {
        username = Encode(prompt("请输入密码"));
        username = HexToDec(username.substring(2,6) + username.substring(8, 12));
        code = ((Date.parse(new Date())/1000) ^ username).toString().padStart(10, "0");
        GetMsg("/signup?key=" + code, function rr(t) {
            j = JSON.parse(t);
            if(j.stat == "success") {
                me = decodeURI(j.id);
                AddUser("uuid", me, 7);
                prompt("这是您的用户名，请复制好后妥善保存", me);
            } else alert("错误!");
        });
    }
}
function Vote(vclass) {
    if(me != "") {
        img = document.getElementById("img_display").src;
        GetMsg("/vote?uuid=" + me + "&img=" + img.substring(img.lastIndexOf('=')+1, img.length) + "&class=" + vclass, function rv(t) {
            GetImg();
        });
    } else alert("请登录!");
}
hidden = false;
col = document.getElementsByTagName("div");
function Show() {
    hidden = !hidden;
    col[0].hidden = col[2].hidden = col[4].hidden = hidden;
    document.getElementById("btn_hide").innerText = hidden?"显示":"隐藏";
}
function Upload() {
    document.getElementById("upload_form").action = "upform?uuid=" + me;
}