// prepare global variables
var volume = 0;
var playerArr = [];
var videoArr = [];
var youtubeQueue = []; //holds YT mediaObjs until IFrameAPI is ready
var ytAPIIsReady = false;
Array.prototype.remove = function(item) {
    itemIndex = this.indexOf(item);
    if (itemIndex > -1) {
        this.splice(itemIndex, 1);
    }
}

// modify prototypes to add and unify functionality
HTMLVideoElement.prototype.spawn = function(mediaObj) {
    this.hide = function() {
        this.setAttribute("hidden", "hidden");
    };
    this.show = function() {
        this.removeAttribute("hidden");
    };
    this.hasStarted = function() {
        videoArr.push(this);
        this.show();
        playerStarted(this.mediaObj.track);
    };
    this.hasEnded = function() {
        if (!this.mediaObj.loop){
            this.hide();
            playerArr.remove(this);
            videoArr.remove(this);
            playerEnded(this.mediaObj.track);
            this.destroy();
        }
    };
    this.hasErrored = function() {
        submessage = "";
        switch (this.error.code) {
            case this.error.MEDIA_ERR_ABORTED:
                submessage = "video file playback has been aborted";
                break;
            case this.error.MEDIA_ERR_NETWORK:
                submessage = "video file download halted due to network error";
                break;
            case this.error.MEDIA_ERR_DECODE:
                submessage = "video file could not be decoded";
                break;
            case this.error.MEDIA_ERR_SRC_NOT_SUPPORTED:
                if (this.networkState == HTMLMediaElement.NETWORK_NO_SOURCE) {
                    submessage = "video file could not be found";
                } else {
                    submessage = "video file format is not supported";
                }
                break;
        }
        playerArr.remove(this);
        videoArr.remove(this);
        playerErrored(this.mediaObj.track, submessage);
        this.destroy();
    };
    this.setVolume = function(volumeLevel) {
        this.volume = volumeLevel / 100.0;
    };
    this.timeUpdated = function() {
        if (this.currentTime >= this.endTime && this.duration != "Inf"){
            if (this.mediaObj.loop){
                this.currentTime = this.startTime;
                this.play();
            } else {
                this.pause();
            }
        }
    }
    this.loadMediaObj = function(mediaObj) {
        this.mediaObj = mediaObj || this.mediaObj;
        var s = this.mediaObj.start;
        var e = this.mediaObj.end;
        if (s.length > 0) {
            this.startTime = s;
        }
        if (e.length > 0) {
            this.endTime = e;
        }
        this.muted = this.mediaObj.muted;
        this.src = this.mediaObj.src + "#t=" + this.startTime;
        if (e.length > 0) {
            this.src += "," + this.endTime;
        }
        this.load();
    };
    this.destroy = function() {
        if (this.parentNode == document.body) {
            document.body.removeChild(this);
        }
    };
    this.ondurationchange = function() {
        this.endTime = this.endTime || this.duration;
        this.hasStarted();
    };
    this.ontimeupdate = function() {this.timeUpdated();};
    this.onerror = function() {this.hasErrored();};
    this.onended =  function() {this.hasEnded();};
    this.onpause = function() {this.hasEnded();};
    this.mediaObj = mediaObj;
    playerArr.push(this);
    this.startTime = 0.0;
    this.endTime = false;
    this.setVolume(volume);
    this.setAttribute("class", "media");
    this.setAttribute("preload", "auto");
    this.setAttribute("autoplay", "autoplay");
    this.loadMediaObj(this.mediaObj);
    document.body.appendChild(this);
    return;
}

HTMLAudioElement.prototype.spawn = function(mediaObj) {
    this.hide = function() {
        this.setAttribute("hidden", "hidden");
    };
    this.show = function() {
        // does nothing to prevent audio player becomming visible
        return;
    };
    this.hasStarted = function() {
        // audio doesn't need to tell parent that it's playing
        //playerStarted(this.mediaObj.track);
    };
    this.hasEnded = function() {
        if (!this.mediaObj.loop){
            playerArr.remove(this);
            playerEnded(this.mediaObj.track);
            this.destroy();
        }
    };
    this.hasErrored = function() {
        submessage = "";
        switch (this.error.code) {
            case this.error.MEDIA_ERR_ABORTED:
                submessage = "audio file playback has been aborted";
                break;
            case this.error.MEDIA_ERR_NETWORK:
                submessage = "audio file download halted due to network error";
                break;
            case this.error.MEDIA_ERR_DECODE:
                submessage = "audio file could not be decoded";
                break;
            case this.error.MEDIA_ERR_SRC_NOT_SUPPORTED:
                if (this.networkState == HTMLMediaElement.NETWORK_NO_SOURCE) {
                    submessage = "audio file could not be found";
                } else {
                    submessage = "audio file format is not supported";
                }
                break;
        }
        playerArr.remove(this);
        playerErrored(this.mediaObj.track, submessage);
        this.destroy();
    };
    this.setVolume = function(volumeLevel) {
        this.volume = volumeLevel / 100.0;
    };
    this.timeUpdated = function() {
        if (this.currentTime > this.endTime && this.duration != "Inf"){
            if (this.mediaObj.loop){
                this.currentTime = this.startTime;
                this.play();
            } else {
                this.pause();
            }
        }
    }
    this.loadMediaObj = function(mediaObj) {
        this.mediaObj = mediaObj || this.mediaObj;
        var s = this.mediaObj.start;
        var e = this.mediaObj.end;
        if (s.length > 0) {
            this.startTime = s;
        }
        if (e.length > 0) {
            this.endTime = e;
        }
        this.muted = this.mediaObj.muted;
        this.src = this.mediaObj.src + "#t=" + this.startTime;
        if (e.length > 0) {
            this.src += "," + this.endTime;
        }
        this.load();
    };
    this.destroy = function() {
        if (this.parentNode == document.body) {
            document.body.removeChild(this);
        }
    };
    this.ondurationchange = function() {
        this.endTime = this.endTime || this.duration;
        this.hasStarted();
    };
    this.ontimeupdate = function() {this.timeUpdated();};
    this.onerror = function() {this.hasErrored();};
    this.onended =  function() {this.hasEnded();};
    this.onpause = function() {this.hasEnded();};
    this.mediaObj = mediaObj;
    playerArr.push(this);
    this.startTime = 0.0;
    this.endTime = false;
    this.shouldLoop = this.mediaObj.loop;
    this.setVolume(volume);
    this.setAttribute("class", "media");
    this.setAttribute("preload", "auto");
    this.setAttribute("autoplay", "autoplay");
    this.loadMediaObj(mediaObj);
    document.body.appendChild(this);
    return;
}

HTMLImageElement.prototype.spawn = function(mediaObj) {
    this.hide = function() {
        this.setAttribute("hidden", "hidden");
    };
    this.show = function() {
        this.removeAttribute("hidden");
        return;
    };
    this.loadMediaObj = function(mediaObj) {
        this.src = mediaObj.src;
    };
    this.destroy = function() {document.body.removeChild(this);};
    playerArr.push(this);
    this.setAttribute("class", "media");
    this.loadMediaObj(mediaObj);
    document.body.appendChild(this);
    return;
}

function onYouTubeIframeAPIReady() {
    // adjust YT.Player prototype for easier management
    YT.Player.prototype.isReady = false;
    YT.Player.prototype.setReady = function() {
        this.setVolume(volume);
        if (this.mediaObj.muted) {
            this.mute();
        }
        this.isReady = true;
        // player may have been given a mediaObj before teh player was ready,
        // so it should now load the desired video with the specified
        // parameters
        this.loadMediaObj();
    };
    // player should hold the mediaObj used to create it, so it can be
    // referenced later
    YT.Player.prototype.mediaObj = false; 
    YT.Player.prototype.pause = function() {
        if (this.isReady) {
            this.pauseVideo();
        }
    };
    YT.Player.prototype.loadMediaObj = function(mediaObj) {
        // if function is being passed a mediaObj, it should override
        // this.mediaObj
        this.mediaObj = mediaObj || this.mediaObj;
        if (this.isReady && this.mediaObj) {
            // if the player is ready AND it has a mediaObj, 
            params = {
                "videoId": this.mediaObj.src,
            }
            var start = this.mediaObj.start;
            var end = this.mediaObj.end;
            if (start.length > 0) {
                params.startSeconds = parseFloat(start);
            } else {
                params.startSeconds = 0;
            }
            if (end.length > 0) {
                params.endSeconds = parseFloat(end);
            }
            this.loadVideoById(params);
        }
    };
    YT.Player.prototype.hide = function() {
        this.getIframe().setAttribute("hidden", "hidden");
    }
    YT.Player.prototype.show = function() {
        this.getIframe().removeAttribute("hidden");
    }
    // once the YouTube Iframe API is ready and the YT.Player prototype has
    // been modified, it will then be safe to start spawning YouTube players
    ytAPIIsReady = true;
    for (ytObj of youtubeQueue) {
        spawnYouTubePlayer(ytObj);
    }
}

function receiveMessage(event) {
    if (event.origin !== "http://localhost:8181" &&
        event.origin !== "https://truveris.com"){
        return;
    }
    var message = JSON.parse(event.data);
    if (message.status == "media") { 
        var mediaObj = message;
        switch (mediaObj.mediaType){
            case "youtube":
                queueYouTubeSpawn(mediaObj);
                break;
            case "video":
                spawnPlayer(mediaObj);
                break;
            case "audio":
                spawnPlayer(mediaObj);
                break;
            case "img":
                spawnPlayer(mediaObj);
                break;
            case "web":
                spawnWeb(mediaObj);
                break;
        }
    }
}

// spawns <video>, <audio>, or <img>
function spawnPlayer(mediaObj) {
    var player = document.createElement(mediaObj.mediaType);
    player.spawn(mediaObj);
    return;
}

// spawns iframe to show webpage
function spawnWeb(mediaObj) {
    var web = document.createElement("iframe");
    web.setAttribute("class", "media");
    web.setAttribute("src", mediaObj.src);
    document.body.appendChild(web);
    return;
}

function queueYouTubeSpawn(mediaObj) {
    // push a new YT mediaObj into the queue, and if the YouTube Iframe API is
    // ready, spawn the players using the mediaObjs in the youtubeQueue array
    youtubeQueue.push(mediaObj);
    if (ytAPIIsReady) {
        for (ytObj of youtubeQueue) {
            spawnYouTubePlayer(ytObj);
        }
    }
}

function spawnYouTubePlayer(mediaObj) {
    // remove the mediaObj from the youtubeQueue array first, so it can't be
    // spawned again
    youtubeQueue.remove(mediaObj);
    // create a <div> for the YouTube player to replace
    var playerDiv = document.createElement("div");
    // use a unique ID so multiple players can be spawned and referenced
    var divId = Math.floor((Math.random() * 100000) + 1).toString();
    playerDiv.setAttribute("id", divId);
    playerDiv.setAttribute("class", "media");
    // hide it at first so it doesn't block anything before it starts actually
    // playing
    playerDiv.setAttribute("hidden", "hidden");
    document.body.appendChild(playerDiv);
    playerParams = {
        height: "100%",
        width: "100%",
        playerVars :{
            "controls": 0,
            "showinfo": 0,
            "rel": 0,
            "modestbranding": 1,
            "iv_load_policy": 3,
            "enablejsapi": 1,
            "origin": "https://truveris.com",
        },
        events: {
            "onReady": onPlayerReady,
            "onStateChange": onPlayerStateChange,
            "onError": onError,
        },
    }
    if (mediaObj.loop == false) {
        // the 'autoplay' parameter messes with looping
        playerParams["playerVars"]["autoplay"] = 1;
    }
    var ytPlayer = new YT.Player(divId, playerParams);
    playerArr.push(ytPlayer);
    ytPlayer.loadMediaObj(mediaObj);
    return;
}

function onPlayerStateChange(event) {
    switch (event.data){
        case YT.PlayerState.UNSTARTED:
            // hide the player so the thumbnail isn't seen while the video
            // isn't playing
            event.target.setPlaybackQuality("highres");
            event.target.hide();
            event.target.playVideo();
            break;
        case YT.PlayerState.PLAYING:
            videoArr.push(event.target);
            // reveal the player now that the thumbnail won't be shown
            event.target.show();
            playerStarted(event.target.mediaObj.track);
            break;
        case YT.PlayerState.ENDED:
            // hide the player so the thumbnail isn't seen while the video
            // isn't playing
            event.target.hide();
            if (event.target.mediaObj.loop) {
                var start = 0;
                if (event.target.mediaObj.start.length > 0) {
                    start = parseFloat(event.target.mediaObj.start);
                }
                event.target.seekTo(start);
                event.target.playVideo();
            } else {
                var divId = event.target.getIframe().getAttribute("id");
                playerArr.remove(event.target);
                videoArr.remove(event.target);
                playerEnded(event.target.mediaObj.track);
                event.target.destroy();
                //remove remaining div
                var containerDiv = document.getElementById(divId);
                if (containerDiv == document.body){
                    document.body.removeChild(containerDiv);
                }
            }
            break;
        case YT.PlayerState.PAUSED:
            // hide the player so the thumbnail isn't seen while the video
            // isn't playing
            event.target.hide();
            if (event.target.mediaObj.loop) {
                event.target.playVideo();
            } else {
                var divId = event.target.getIframe().getAttribute("id");
                playerArr.remove(event.target);
                videoArr.remove(event.target);
                playerEnded(event.target.mediaObj.track);
                event.target.destroy();
                //remove remaining div
                var containerDiv = document.getElementById(divId);
                if (containerDiv.parentNode == document.body){
                    document.body.removeChild(containerDiv);
                }
            }
            break;
    }
    return;
}

function onPlayerReady(event) {
    // YouTube player is now ready
    event.target.setReady();
}

function onError(event) {
    var srcTrack = event.target.mediaObj.track;
    submessage = "";
    switch(event.data){
        case 2:
            submessage =  "invalid youtube video parameter"
            break;
        case 5:
            submessage =  "youtube video doesn't work with html5"
            break;
        case 100:
            submessage =  "no such youtube video"
            break;
        case 101:
            submessage =  "can't embed this youtube video"
            break;
        case 150:
            submessage =  "can't embed this youtube video"
            break;
    }
    submessage = submessage || "";
    // remove all traces of the player
    var divId = event.target.getIframe().getAttribute("id");
    playerArr.remove(event.target);
    videoArr.remove(event.target);
    playerEnded(event.target.mediaObj.track);
    event.target.destroy();
    //remove remaining div
    var containerDiv = document.getElementById(divId);
    if (containerDiv == document.body){
        document.body.removeChild(containerDiv);
    }
    playerErrored(srcTrack, submessage);
    return;
}

function playerStarted(srcTrack) {
    sendMessage(srcTrack, "PLAYING");
    return;
}

function playerEnded(srcTrack) {
    if (videoArr.length == 0) {
        sendMessage(srcTrack, "ENDED");
    }
    return;
}

function playerErrored(srcTrack, submessage) {
    submessage = submessage || "";
    sendMessage(srcTrack, "ERRORED", submessage);
    playerEnded(srcTrack);
    return;
}

function sendMessage(srcTrack, state, submessage) {
    message = {
        source: srcTrack,
        playerState: state,
        submessage: submessage,
    }
    parent.postMessage(JSON.stringify(message), "*");
    return;
}

function setVolume(newVolume) {
    volume = newVolume;
    for (player of playerArr) {
        if (player.setVolume){
            player.setVolume(volume);
        }
    }
}

function shutup() {
    // kills all the players
    while (document.body.children.length > 0) {
        document.body.removeChild(document.body.children[0]);
    }
    // it's safer to clear the arrays after body is bodied
    playerArr = [];
    videoArr = [];
    youtubeQueue = [];
}

window.onload=function(){
    // handle messages from parent and children
    window.addEventListener("message", receiveMessage, false);
    // set the volume variable to parent window's volume variable
    volume = parent.volume * 100;
    // Load the IFrame Player API code asynchronously
    var tag = document.createElement('script');
    tag.src = "https://www.youtube.com/iframe_api";
    var firstScriptTag = document.getElementsByTagName('script')[0];
    firstScriptTag.parentNode.insertBefore(tag, firstScriptTag);
    return;
}
