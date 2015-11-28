// prepare global variables
var volume = 0; // master volume (0 by default)
var trackVolume = 1.0; // percentage of master volume to use in this track

// modify prototypes to add and unify functionality
function modifyMediaElementPrototypes() {
    HTMLMediaElement.prototype.hide = function() {
        this.setAttribute("opacity", 0);
    };
    HTMLVideoElement.prototype.show = function() {
        playerStarted();
        this.setAttribute("opacity", 1);
    };
    HTMLAudioElement.prototype.show = function() {
        // does nothing to prevent audio player becomming visible
        return;
    };
    HTMLVideoElement.prototype.hasStarted = function() {
        this.show();
    };
    HTMLAudioElement.prototype.hasStarted = function() {
        // audio doesn't need to tell parent that it's playing
    };
    HTMLMediaElement.prototype.hasEnded = function() {
        if (this.didEnd == false){
            this.didEnd = true;
            if (!this.soloLoop){
                this.hide();
                playerEnded();
                this.destroy();
            }
        }
    };
    HTMLMediaElement.prototype.hasErrored = function(event) {
        submessage = "";
        type = this.nodeName.toLowerCase();
        switch (this.error.code) {
            case this.error.MEDIA_ERR_ABORTED:
                err = " file playback has been aborted";
                break;
            case this.error.MEDIA_ERR_NETWORK:
                err = " file download halted due to network error";
                break;
            case this.error.MEDIA_ERR_DECODE:
                err = " file could not be decoded";
                break;
            case this.error.MEDIA_ERR_SRC_NOT_SUPPORTED:
                if (this.networkState == HTMLMediaElement.NETWORK_NO_SOURCE) {
                    err = " file could not be found";
                } else {
                    err = " file format is not supported";
                }
                break;
        }
        submessage = type + err
        reportError(submessage);
        this.hasEnded()
    };
    HTMLMediaElement.prototype.setVolume = function(volumeLevel) {
        this.volume = volumeLevel / 100.0;
    };
    HTMLMediaElement.prototype.timeUpdated = function(event) {
        if (this.currentTime >= this.endTime &&
                this.duration != "Inf" &&
                !this.didEnd){
            if (this.soloLoop){
                this.currentTime = 0;
                this.play();
            } else {
                this.hide();
                this.pause();
                this.hasEnded();
            }
        }
    };
    HTMLMediaElement.prototype.loadMediaObj = function() {
        var e = this.mediaObj.end;
        if (e.length > 0) {
            this.endTime = e;
        }
        this.muted = this.mediaObj.muted;
        this.src = this.mediaObj.src;
    };
    HTMLMediaElement.prototype.seekToEnd = function() {
        this.currentTime = this.endTime;
    };
    HTMLMediaElement.prototype.destroy = function() {
        this.parentNode.removeChild(this);
    };
    HTMLMediaElement.prototype.spawn = function(mediaObj) {
        this.ondurationchange = function() {
            this.endTime = this.endTime || this.duration;
            this.hasStarted();
        };
        this.onplay = function() {
            this.show();
            this.didEnd = false;
        };
        this.ontimeupdate = function(event) {this.timeUpdated(event);};
        this.onerror = function(event) {this.hasErrored(event);};
        this.onended =  function() {this.hasEnded();};
        //this.onpause = function() {this.hasEnded();};
        this.mediaObj = mediaObj;
        this.soloLoop = mediaObj.loop;
        this.endTime = false;
        this.didEnd = false;
        this.setVolume(volume * trackVolume);
        this.loadMediaObj();
        this.setAttribute("class", "media");
        this.setAttribute("opacity", 0);
        this.setAttribute("preload", "auto");
        document.body.appendChild(this);
        this.play();
        return;
    };
}

function modifyImgElementPrototype() {
    HTMLImageElement.prototype.hide = function() {
        this.setAttribute("opacity", 0);
    };
    HTMLImageElement.prototype.show = function() {
        playerStarted();
        this.setAttribute("opacity", 1);
    };
    HTMLImageElement.prototype.loadMediaObj = function() {
        this.src = this.mediaObj.src;
    };
    HTMLImageElement.prototype.destroy = function() {
        this.parentNode.removeChild(this);
    };
    HTMLImageElement.prototype.setVolume = function(volumeLevel) {
        // should do nothing
        return;
    };
    HTMLImageElement.prototype.seekToEnd = function(volumeLevel) {
        return;
    };
    HTMLImageElement.prototype.spawn = function(mediaObj) {
        this.mediaObj = mediaObj;
        this.setAttribute("class", "media");
        this.loadMediaObj();
        document.body.appendChild(this);
        return;
    }
    return;
}

function modifyIframeElementPrototype() {
    HTMLIFrameElement.prototype.hide = function() {
        this.setAttribute("opacity", 0);
    };
    HTMLIFrameElement.prototype.show = function() {
        playerStarted();
        this.setAttribute("opacity", 1);
    };
    HTMLIFrameElement.prototype.loadMediaObj = function() {
        this.src = this.mediaObj.src;
    };
    HTMLIFrameElement.prototype.destroy = function() {
        this.parentNode.removeChild(this);
    };
    HTMLIFrameElement.prototype.setVolume = function(volumeLevel) {
        // if the iframe is being used to embed a player, call the player's
        // setVolume function.
        if(this.player) {
            this.player.setVolume(volumeLevel);
        }
        return;
    };
    HTMLIFrameElement.prototype.seekToEnd = function(volumeLevel) {
        return;
    };
    HTMLIFrameElement.prototype.spawn = function(mediaObj) {
        this.mediaObj = mediaObj;
        this.setAttribute("class", "media");
        this.loadMediaObj();
        document.body.appendChild(this);
        return;
    }
    return;
}

/* -------------------------------- YOUTUBE -------------------------------- */

function modifyYouTubePlayerPrototype() {
    // adjust YT.Player prototype for easier management
    YT.Player.prototype.isReady = false;
    YT.Player.prototype.startTime = 0.0;
    YT.Player.prototype.endTime = false;
    YT.Player.prototype.setReady = function() {
        this.setVolume(volume * trackVolume);
        iframe = this.getIframe();
        iframe.player = this;
        if (this.mediaObj.muted) {
            this.mute();
        }
        this.isReady = true;
        // player may have been given a mediaObj before the player was ready,
        // so it should now load the desired video with the specified
        // parameters
        if (this.mediaObj) {
            this.loadMediaObj();
        }
    };
    // player should hold the mediaObj used to create it, so it can be
    // referenced later
    YT.Player.prototype.mediaObj = false;
    YT.Player.prototype.pause = function() {
        if (this.isReady) {
            this.pauseVideo();
        }
    };
    YT.Player.prototype.play = function() {
        if (this.isReady) {
            this.playVideo();
        }
    };
    YT.Player.prototype.seekToStart = function() {
        this.seekTo(this.startTime);
    };
    YT.Player.prototype.seekToEnd = function() {
        this.endTime = this.endTime || this.getDuration();
        this.seekTo(this.endTime);
    };
    YT.Player.prototype.destroy = function() {
        iframe = this.getIframe();
        iframe.destroy();
    };
    YT.Player.prototype.loadMediaObj = function() {
        if (this.isReady) {
            // if the player is ready
            params = {
                "videoId": this.mediaObj.src,
            }
            var end = this.mediaObj.end;
            if (end.length > 0) {
                params.endSeconds = parseFloat(end);
                this.endTime = end;
            }
            this.soloLoop = mediaObj.loop;
            this.loadVideoById(params);
        }
    };
    YT.Player.prototype.hide = function() {
        this.getIframe().setAttribute("opacity", 0);
    };
    YT.Player.prototype.show = function() {
        playerStarted();
        this.getIframe().setAttribute("opacity", 1);
    };
    YT.Player.prototype.containerId = null;
    YT.Player.prototype.soloLoop = false;
}

function spawnYouTubePlayer(mediaObj) {
    // create a <div> for the YouTube player to replace
    var playerDiv = document.createElement("div");
    // use a unique ID so multiple players can be spawned and referenced
    var containerId = Math.floor((Math.random() * 100000) + 1).toString();
    playerDiv.setAttribute("id", containerId);
    playerDiv.setAttribute("class", "media");
    // hide it at first so it doesn't block anything before it starts actually
    // playing
    playerDiv.setAttribute("opacity", 0);
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
            "onReady": onYTPlayerReady,
            "onStateChange": onYTPlayerStateChange,
            "onError": onYTPlayerError,
        },
    }
    var ytPlayer = new YT.Player(containerId, playerParams);
    ytPlayer.containerId = containerId;
    ytPlayer.mediaObj = mediaObj;
    ytPlayer.loadMediaObj();

    return ytPlayer;
}

//Embedded players' state change handlers
function onYTPlayerStateChange(event) {
    switch (event.data){
        case YT.PlayerState.UNSTARTED:
            // hide the player so the thumbnail isn't seen while the video
            // isn't playing

            event.target.setPlaybackQuality("highres");
            event.target.hide();
            event.target.playVideo();
            break;
        case YT.PlayerState.PLAYING:
            // reveal the player now that the thumbnail won't be shown
            event.target.show();
            break;
        case YT.PlayerState.ENDED:
            // hide the player so the thumbnail isn't seen while the video
            // isn't playing
            event.target.hide();
            event.target.seekToStart();
            if (event.target.soloLoop) {
                event.target.playVideo();
            } else {
                playerEnded();
                event.target.destroy();
            }
            break;
    }
    return;
}

//Embedded players' ready state handlers
function onYTPlayerReady(event) {
    // YouTube player is now ready
    event.target.setReady();
}

//Embedded players' error handling
function onYTPlayerError(event) {
    submessage = "";
    switch(event.data){
        case 2:
            submessage = "invalid youtube video parameter"
            break;
        case 5:
            submessage = "youtube video doesn't work with html5"
            break;
        case 100:
            submessage = "no such youtube video"
            break;
        case 101:
            submessage = "can't embed this youtube video"
            break;
        case 150:
            submessage = "can't embed this youtube video"
            break;
    }
    submessage = submessage || "unrecognized error code: " + event.data;
    // remove all traces of the player
    reportError(this.miniPlaylist.track, submessage);
    playerEnded();
    this.destroy();
    return;
}

/* ------------------------------ END YOUTUBE ------------------------------ */

function receiveMessage(event) {
    if (event.origin !== "http://localhost:8181" &&
        event.origin !== "https://truveris.com"){
        return;
    }
    var message = JSON.parse(event.data);
    if (message.status == "media") {
        mediaObj = message.mediaObj;
        spawnMediaObj(mediaObj);
    }
}

function spawnMediaObj(mediaObj) {
    switch (mediaObj.format){
        case "video":
            spawnStandardPlayer(mediaObj);
            break;
        case "audio":
            spawnStandardPlayer(mediaObj);
            break;
        case "img":
            spawnStandardPlayer(mediaObj);
            break;
        case "web":
            spawnWeb(mediaObj);
            break;
        case "youtube":
            spawnYouTubePlayer(mediaObj);
            break;
        default:
            reportError("unrecognized format: " + mediaObj.format)
    }
}

// spawns <video>, <audio>, or <img>
function spawnStandardPlayer(mediaObj) {
    var player = document.createElement(mediaObj.format);
    player.spawn(mediaObj);
    return player;
}

// spawns iframe to show webpage
function spawnWeb(mediaObj) {
    var web = document.createElement("iframe");
    web.spawn(mediaObj);
    return web;
}

function playerStarted() {
    sendMessage("PLAYING");
    return;
}

function playerEnded() {
    if (getVisibleCount() == 0) {
        sendMessage("ENDED");
    }
    return;
}

function getVisibleCount() {
    count = 0;
    for(el of getAllPlayers()) {
        if(el.getAttribute("opacity") != 0) {
            count++;
        }
    }
    return count
}

function getAllPlayers() {
    collection = document.querySelectorAll("body > *");
    elementArr = [];
    for(i = 0; i < collection.length; i++){
        elementArr.push(collection.item(i));
    }
    return elementArr;
}

function reportError(submessage) {
    submessage = submessage || "";
    sendMessage("ERRORED", submessage);
    return;
}

function sendMessage(state, submessage) {
    message = {
        playerState: state,
        submessage: submessage,
    }
    parent.postMessage(JSON.stringify(message), "*");
    return;
}

function setVolume(newVolume) {
    if (newVolume !== undefined && newVolume !== null){
        volume = newVolume;
    }
    for (player of getAllPlayers()) {
        if (player.setVolume) {
            player.setVolume(volume * trackVolume);
        }
    }
    return;
}

function setTrackVolume(newVolume) {
    trackVolume = newVolume;
    setVolume();
    return;
}

function shutup() {
    // kills all the players
    while (getAllPlayers().length > 0){
        getAllPlayers()[0].destroy();
    }
    playerEnded();
    return;
}

window.onload=function(){
    // handle messages from parent and children
    window.addEventListener("message", receiveMessage, false);
    // set the volume variable to parent window's volume variable
    volume = parent.volume;

    modifyMediaElementPrototypes();
    modifyImgElementPrototype();
    modifyIframeElementPrototype();
    modifyYouTubePlayerPrototype();

    return;
}
