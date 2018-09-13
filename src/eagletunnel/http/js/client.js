function changeProxyStatus(){
    var status = document.getElementById("proxy-status");
    var index = status.selectedIndex;
    var value = status.options[index].value;
    var list = document.getElementById("whitelist_domains_label");
    if(value =="全局"){
        list.style.display = "none";
    }else{
        list.style.display = "inline";
    }
}

function changeUserCheck(){
    var uc = document.getElementById("user-check");
    var index = uc.selectedIndex;
    var value = uc.options[index].value;
    var id_label = document.getElementById("id_label");
    var password_label = document.getElementById("password_label");
    if(value == "开启"){
        id_label.style.display = "inline";
        password_label.style.display = "inline";
    }else{
        id_label.style.display = "none";
        password_label.style.display = "none";
    }
}

function init(){
    changeProxyStatus();
    changeUserCheck();
}