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

function init(){
    changeProxyStatus();
}