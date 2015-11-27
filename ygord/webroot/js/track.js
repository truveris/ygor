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
        this.endTime = "";
        if (e.length > 0) {
            this.endTime = e;
        }
        this.muted = this.mediaObj.muted;
        boundries = "#t=0";
        if (this.endTime.length > 0){
            boundries += "," + this.endTime;
        }
        this.src = this.mediaObj.src + boundries;
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
        this.setAttribute("class", "media");
        this.setAttribute("opacity", 0);
        document.body.appendChild(this);
        this.loadMediaObj();
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
    count = 0
    for(el of getAllPlayers()) {
        if(el.getAttribute("opacity") != 0) {
            count++;
        }
    }
    return count
}

function getAllPlayers() {
    return document.querySelectorAll("body > *");
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
    return;
}

window.onload=function(){
    // handle messages from parent and children
    window.addEventListener("message", receiveMessage, false);
    // set the volume variable to parent window's volume variable
    volume = parent.volume;

    modifyMediaElementPrototypes();
    modifyImgElementPrototype();

    return;
}
