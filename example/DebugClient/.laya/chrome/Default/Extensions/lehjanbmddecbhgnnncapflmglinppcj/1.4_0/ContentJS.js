
chrome.extension.onRequest.addListener(function(request, sender, sendResponse) {

    sendResponse(1);

});



var sendMsg = function() {
	
    var host = document.getElementById("ClCache").attributes["host"].value;

    host = host+"|thatsall";
	
    chrome.extension.sendRequest({ msg: host }, function(results) {  });
	
};


var objdiv = document.createElement("div");

objdiv.innerHTML = "<object id=\"ClCache\" click=\"sendMsg\" host=\"\" width=0 height=0></object>";

document.body.appendChild(objdiv);

document.getElementById("ClCache").addEventListener('click', sendMsg, false);

	



