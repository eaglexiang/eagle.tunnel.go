function changeUserCheck(){
    var user_check = document.getElementById("user-check");
    var index = user_check.selectedIndex;
    var value = user_check.options[index].value;
    var users = document.getElementById("users_label");
    if(value =="关闭"){
        users.style.display = "none";
    }else{
        users.style.display = "inline";
    }
}

function init(){
    changeUserCheck();
}