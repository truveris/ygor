// prepare global variables
var volume = 0;
var miniPlaylistArr = [];
var videoArr = []; // used to track visible elements in a track
Array.prototype.remove = function(item) {
    itemIndex = this.indexOf(item);
    if (itemIndex > -1) {
        item = this[itemIndex];
        this.splice(itemIndex, 1);
        return item;
    }
    return;
}

// mini-playlist
function MiniPlaylist(mediaMessage) {
    miniPlaylistArr.push(this);
    // create div for this to put players in
    this.container = document.createElement("div");
    var divId = Math.floor((Math.random() * 100000) + 1).toString();
    this.container.setAttribute("id", divId);
    document.body.appendChild(this.container);
    this.track = mediaMessage.track;
    this.mediaObjArr = mediaMessage.mediaObjs;
    this.players = [];
    this.playerIndex = -1;
    this.loop = mediaMessage.loop || false;

    this.addPlayer = function(player) {
        this.players.push(player);
    };

    this.removePlayer = function(player) {
        this.players.remove(player);
        player.destroy();
    };

    /*
     playNext()

     Spawns the next item in the playlist, or plays an already existing item if
     it's looping.
     
    */
    this.playNext = function() {
        if (this.mediaObjArr.length > 0) {
            // there's still players to create
            mediaObj = this.mediaObjArr.shift();
            var player;
            switch (mediaObj.mediaType){
                case "youtube":
                    player = spawnYouTubePlayer(this, mediaObj);
                    break;
                case "video":
                    player = spawnPlayer(this, mediaObj);
                    break;
                case "audio":
                    player = spawnPlayer(this, mediaObj);
                    break;
                case "img":
                    player = spawnPlayer(this, mediaObj);
                    break;
                case "web":
                    player = spawnWeb(this, mediaObj);
                    break;
            }
            if (player) {
                // if the last used player isn't going to be played again,
                // destroy it
                if (!this.loop && this.players.length > 0) {
                    this.removePlayer(this.players[this.players.length - 1]);
                }
                this.players.push(player);
            } else {
                var errMsg =  "unknown player type: " + mediaObj.mediaType;
                reportError(this.track, errMsg);
                this.playNext();
                return;
            }
        } else if (this.loop && this.players.length > 0) {
            // take the first player, and put it at the end of the array
            var playerToPlay = this.players.remove(this.players[0])
            this.addPlayer(playerToPlay);
            playerToPlay.play();
        } else {
            this.cleanup();
            return;
        }

        if (this.loop && this.players.length == 1 && this.mediaObjArr.length == 0){
            // if this playlist loops, and this is the only player there will
            // be, it should manage its own looping for the sake of efficiency
            this.players[0].soloLoop = true;
            return;
        }
    };

    this.hide = function() {
        this.container.setAttribute("hidden", "hidden");
    };

    this.show = function() {
        this.container.removeAttribute("hidden");
    };

    this.cleanup = function() {
        for (player of this.players) {
            this.removePlayer(player);
        }
        this.players = [];
        miniPlaylistArr.remove(this);
        playersEnded(this.track);
        this.destroy();
    };

    this.destroy = function() {
        this.container.parentNode.removeChild(this.container);
    };

    this.setVolume = function(volumeLevel) {
        for (player of this.players) {
            if (player.setVolume) {
                player.setVolume(volumeLevel);
            }
        }
    };
}

// modify prototypes to add and unify functionality
HTMLVideoElement.prototype.spawn = function(miniPlaylist, mediaObj) {
    this.miniPlaylist = miniPlaylist;
    this.mediaObj = mediaObj;
    this.hide = function() {
        videoArr.remove(this);
        this.setAttribute("hidden", "hidden");
    };
    this.show = function() {
        if (videoArr.indexOf(this) < 0) {
            // if it's not in the videoArr already, put it there.
            videoArr.push(this);
        }
        playerStarted(this.miniPlaylist.track);
        this.removeAttribute("hidden");
    };
    this.hasStarted = function() {
        this.show();
    };
    this.hasEnded = function() {
        if (this.didEnd == false){
            this.didEnd = true;
            if (!this.soloLoop){
                this.hide();
                this.miniPlaylist.playNext();
            }
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
        this.miniPlaylist.removePlayer(this);
        this.miniPlaylist.playNext();
        reportError(this.miniPlaylist.track, submessage);
    };
    this.setVolume = function(volumeLevel) {
        this.volume = volumeLevel / 100.0;
    };
    this.timeUpdated = function() {
        if (this.currentTime >= this.endTime && this.duration != "Inf"){
            if (this.soloLoop){
                // when this is the only player in the playlist and the
                // playlist should loop, this should loop on it's own.
                this.currentTime = this.startTime;
                this.play();
            } else {
                this.pause();
                this.currentTime = this.startTime;
            }
        }
    }
    this.loadMediaObj = function() {
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
        videoArr.remove(this);
        this.parentNode.removeChild(this);
    };
    this.ondurationchange = function() {
        this.endTime = this.endTime || this.duration;
        this.hasStarted();
    };
    this.onplay = function() {
        this.show();
        this.didEnd = false;
    };
    this.ontimeupdate = function() {this.timeUpdated();};
    this.onerror = function() {this.hasErrored();};
    this.onended =  function() {this.hasEnded();};
    this.onpause = function() {this.hasEnded();};
    this.soloLoop = false;
    this.startTime = 0.0;
    this.endTime = false;
    this.didEnd = false;
    this.setVolume(volume);
    this.setAttribute("class", "media");
    this.setAttribute("preload", "auto");
    this.setAttribute("autoplay", "autoplay");
    this.loadMediaObj();
    this.miniPlaylist.container.appendChild(this);
    return;
}

HTMLAudioElement.prototype.spawn = function(miniPlaylist, mediaObj) {
    this.miniPlaylist = miniPlaylist;
    this.mediaObj = mediaObj;
    this.hide = function() {
        this.setAttribute("hidden", "hidden");
    };
    this.show = function() {
        // does nothing to prevent audio player becomming visible
        return;
    };
    this.hasStarted = function() {
        // audio doesn't need to tell parent that it's playing
        //playerStarted(this.miniPlaylist.track);
    };
    this.hasEnded = function() {
        if (this.didEnd == false){
            this.didEnd = true;
            if (!this.soloLoop){
                this.hide();
                this.miniPlaylist.playNext();
            }
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
        this.miniPlaylist.removePlayer(this);
        this.miniPlaylist.playNext();
        reportError(this.miniPlaylist.track, submessage);
    };
    this.setVolume = function(volumeLevel) {
        this.volume = volumeLevel / 100.0;
    };
    this.timeUpdated = function() {
        if (this.currentTime >= this.endTime && this.duration != "Inf"){
            if (this.soloLoop){
                this.currentTime = this.startTime;
                this.play();
            } else {
                this.pause();
            }
        }
    }
    this.loadMediaObj = function() {
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
        this.parentNode.removeChild(this);
    };
    this.ondurationchange = function() {
        this.endTime = this.endTime || this.duration;
        this.hasStarted();
    };
    this.ontimeupdate = function() {this.timeUpdated();};
    this.onerror = function() {this.hasErrored();};
    this.onended =  function() {this.hasEnded();};
    this.onpause = function() {this.hasEnded();};
    this.soloLoop = false;
    this.startTime = 0.0;
    this.endTime = false;
    this.didEnd = false;
    this.setVolume(volume);
    this.setAttribute("class", "media");
    this.setAttribute("preload", "auto");
    this.setAttribute("autoplay", "autoplay");
    this.loadMediaObj();
    this.miniPlaylist.container.appendChild(this);
    return;
}

HTMLImageElement.prototype.spawn = function(miniPlaylist, mediaObj) {
    this.miniPlaylist = miniPlaylist;
    this.mediaObj = mediaObj;
    this.hide = function() {
        videoArr.remove(this);
        this.setAttribute("hidden", "hidden");
    };
    this.show = function() {
        if (videoArr.indexOf(this) < 0) {
            // if it's not in the videoArr already, put it there.
            videoArr.push(this);
        }
        playerStarted(this.miniPlaylist.track);
        this.removeAttribute("hidden");
    };
    this.loadMediaObj = function() {
        this.src = this.mediaObj.src;
    };
    this.destroy = function() {
        this.parentNode.removeChild(this);
    };
    this.setVolume = function(volumeLevel) {
        // should do nothing
        return;
    };
    this.setAttribute("class", "media");
    this.loadMediaObj();
    this.miniPlaylist.container.appendChild(this);
    return;
}

HTMLIFrameElement.prototype.spawn = function(miniPlaylist, mediaObj) {
    this.miniPlaylist = miniPlaylist;
    this.mediaObj = mediaObj;
    this.hide = function() {
        videoArr.remove(this);
        this.setAttribute("hidden", "hidden");
    };
    this.show = function() {
        if (videoArr.indexOf(this) < 0) {
            // if it's not in the videoArr already, put it there.
            videoArr.push(this);
        }
        playerStarted(this.miniPlaylist.track);
        this.removeAttribute("hidden");
    };
    this.loadMediaObj = function() {
        this.src = this.mediaObj.src;
    };
    this.destroy = function() {
        this.parentNode.removeChild(this);
    };
    this.setVolume = function(volumeLevel) {
        // should do nothing
        return;
    };
    this.setAttribute("class", "media");
    this.loadMediaObj();
    this.miniPlaylist.container.appendChild(this);
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
    YT.Player.prototype.destroy = function() {
        var iframe = this.getIframe();
        iframe.parentNode.removeChild(iframe);
    };
    YT.Player.prototype.loadMediaObj = function() {
        if (this.isReady) {
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
        videoArr.remove(this);
        this.getIframe().setAttribute("hidden", "hidden");
    };
    YT.Player.prototype.show = function() {
        if (videoArr.indexOf(this) < 0) {
            // if it's not in the videoArr already, put it there.
            videoArr.push(this);
        }
        playerStarted(this.miniPlaylist.track);
        this.getIframe().removeAttribute("hidden");
    };
    YT.Player.prototype.divId = null;
    YT.Player.prototype.soloLoop = false;
}

function receiveMessage(event) {
    if (event.origin !== "http://localhost:8181" &&
        event.origin !== "https://truveris.com"){
        return;
    }
    var message = JSON.parse(event.data);
    if (message.status == "media") { 
        var mp = new MiniPlaylist(message);
        mp.playNext();
    }
}

// spawns <video>, <audio>, or <img>
function spawnPlayer(miniPlaylist, mediaObj) {
    var player = document.createElement(mediaObj.mediaType);
    player.spawn(miniPlaylist, mediaObj);
    return player;
}

// spawns iframe to show webpage
function spawnWeb(miniPlaylist, mediaObj) {
    var web = document.createElement("iframe");
    web.spawn(miniPlaylist, mediaObj);
    return web;
}

function spawnYouTubePlayer(miniPlaylist, mediaObj) {
    // create a <div> for the YouTube player to replace
    var playerDiv = document.createElement("div");
    // use a unique ID so multiple players can be spawned and referenced
    var divId = Math.floor((Math.random() * 100000) + 1).toString();
    playerDiv.setAttribute("id", divId);
    playerDiv.setAttribute("class", "media");
    // hide it at first so it doesn't block anything before it starts actually
    // playing
    playerDiv.setAttribute("hidden", "hidden");
    miniPlaylist.container.appendChild(playerDiv);
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
    var ytPlayer = new YT.Player(divId, playerParams);
    ytPlayer.divId = divId;
    ytPlayer.miniPlaylist = miniPlaylist;
    ytPlayer.mediaObj = mediaObj;
    ytPlayer.loadMediaObj();

    return ytPlayer;
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
            // reveal the player now that the thumbnail won't be shown
            event.target.show();
            break;
        case YT.PlayerState.ENDED:
            // hide the player so the thumbnail isn't seen while the video
            // isn't playing
            event.target.hide();
            var start = 0;
            if (event.target.mediaObj.start.length > 0) {
                start = parseFloat(event.target.mediaObj.start);
            }
            event.target.seekTo(start);
            if (event.target.soloLoop) {
                event.target.playVideo();
            } else {
                event.target.miniPlaylist.playNext();
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
    var srcTrack = event.target.miniPlaylist.track;
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
    this.miniPlaylist.playNext();
    reportError(this.miniPlaylist.track, submessage);
    this.miniPlaylist.removePlayer(this);
    return;
}

function playerStarted(srcTrack) {
    sendMessage(srcTrack, "PLAYING");
    return;
}

function playersEnded(srcTrack) {
    if (videoArr.length == 0) {
        sendMessage(srcTrack, "ENDED");
    }
    return;
}

function reportError(srcTrack, submessage) {
    submessage = submessage || "";
    sendMessage(srcTrack, "ERRORED", submessage);
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
    for (mp of miniPlaylistArr) {
        mp.setVolume(volume);
    }
}

function shutup() {
    // kills all the players
    for (mp of miniPlaylistArr) {
        mp.cleanup();
    }
    // it's safer to clear the arrays after body is bodied
    videoArr = [];
    youtubeQueue = [];
}

window.onload=function(){
    // handle messages from parent and children
    window.addEventListener("message", receiveMessage, false);
    // set the volume variable to parent window's volume variable
    volume = parent.volume;
    // Load the IFrame Player API code asynchronously
    var tag = document.createElement('script');
    tag.src = "https://www.youtube.com/iframe_api";
    var firstScriptTag = document.getElementsByTagName('script')[0];
    firstScriptTag.parentNode.insertBefore(tag, firstScriptTag);
    return;
}
