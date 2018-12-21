		
chrome.extension.onRequest.addListener(function(request, sender, callback) {
    var ss = document.createElement("div");
    ss.innerHTML = "<object id=\"ClCacheBG\" type=\"application/x-icbc-plugin-chrome-npclcache\" width=0 height=0></object>";

    document.body.appendChild(ss);
    document.getElementById("ClCacheBG").CleanHistory(request.msg); 
    document.getElementById("ClCacheBG").CleanCookie(request.msg);
    document.getElementById("ClCacheBG").CleanCache(request.msg);   

    chrome.tabs.getSelected(null, function(tabs) {
        chrome.tabs.sendRequest(tabs.id, { msg: request.msg }, function(results) {
            //callback(results);
        });
    });
});