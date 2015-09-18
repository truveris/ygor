// prepare global variables
var volume = 0; // master volume (0 by default)
var trackVolume = 1.0; // percentage of master volume to use in this track
var miniPlaylistArr = [];
var videoArr = []; // used to track visible elements in a track
var soundcloudClientId = "1b10b405b1065aadc1639a2620521638";

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
    var containerId = Math.floor((Math.random() * 100000) + 1).toString();
    this.container.setAttribute("id", containerId);
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
                case "soundcloud":
                    player = spawnSoundCloudPlayer(this, mediaObj);
                    break;
                case "vimeo":
                    player = spawnVimeoPlayer(this, mediaObj);
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
        while (this.players.length > 0){
            this.removePlayer(this.players[0]);
        }
        this.players = [];
        miniPlaylistArr.remove(this);
        playersEnded(this.track);
        this.destroy();
    };

    this.destroy = function() {
        if (this.container.parentNode) {
            this.container.parentNode.removeChild(this.container);
        }
    };

    this.setVolume = function(volumeLevel) {
        for (player of this.players) {
            if (player.setVolume) {
                player.setVolume(volumeLevel);
            }
        }
    };

    this.skip = function() {
        currentPlayer = this.players[this.players.length - 1];
        if (currentPlayer.seekToEnd) {
            currentPlayer.seekToEnd();
        }
    };
}

// modify prototypes to add and unify functionality
function modifyMediaElementPrototypes() {
    HTMLMediaElement.prototype.hide = function() {
        this.setAttribute("hidden", "hidden");
    };
    HTMLVideoElement.prototype.show = function() {
        if (videoArr.indexOf(this) < 0) {
            // if it's not in the videoArr already, put it there.
            videoArr.push(this);
        }
        playerStarted(this.miniPlaylist.track);
        this.removeAttribute("hidden");
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
        //playerStarted(this.miniPlaylist.track);
    };
    HTMLMediaElement.prototype.hasEnded = function() {
        if (this.didEnd == false){
            this.didEnd = true;
            if (!this.soloLoop){
                this.hide();
                this.miniPlaylist.playNext();
            }
        }
    };
    HTMLMediaElement.prototype.hasErrored = function() {
        submessage = "";
        type = this.nodeName.toLowerCase();
        switch (this.error.code) {
            case this.error.MEDIA_ERR_ABORTED:
                submessage = type + " file playback has been aborted";
                break;
            case this.error.MEDIA_ERR_NETWORK:
                submessage = type + " file download halted due to network error";
                break;
            case this.error.MEDIA_ERR_DECODE:
                submessage = type + " file could not be decoded";
                break;
            case this.error.MEDIA_ERR_SRC_NOT_SUPPORTED:
                if (this.networkState == HTMLMediaElement.NETWORK_NO_SOURCE) {
                    submessage = type + " file could not be found";
                } else {
                    submessage = type + " file format is not supported";
                }
                break;
        }
        this.miniPlaylist.removePlayer(this);
        this.miniPlaylist.playNext();
        reportError(this.miniPlaylist.track, submessage);
    };
    HTMLMediaElement.prototype.setVolume = function(volumeLevel) {
        this.volume = volumeLevel / 100.0;
    };
    HTMLMediaElement.prototype.timeUpdated = function() {
        if (this.currentTime >= this.endTime && this.duration != "Inf"){
            if (this.soloLoop){
                // when this is the only player in the playlist and the
                // playlist should loop, this should loop on it's own.
                this.currentTime = this.startTime;
                this.play();
            } else {
                this.pause();
                this.hasEnded();
                this.currentTime = this.startTime;
            }
        }
    };
    HTMLMediaElement.prototype.loadMediaObj = function() {
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
    HTMLMediaElement.prototype.seekToEnd = function() {
        this.currentTime = this.endTime;
    };
    HTMLMediaElement.prototype.destroy = function() {
        videoArr.remove(this);
        this.parentNode.removeChild(this);
    };    
    HTMLMediaElement.prototype.spawn = function(miniPlaylist, mediaObj) {
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
        //this.onpause = function() {this.hasEnded();};
        this.miniPlaylist = miniPlaylist;
        this.mediaObj = mediaObj;
        this.soloLoop = false;
        this.startTime = 0.0;
        this.endTime = false;
        this.didEnd = false;
        this.setVolume(volume * trackVolume);
        this.setAttribute("class", "media");
        this.setAttribute("preload", "auto");
        this.setAttribute("autoplay", "autoplay");
        this.loadMediaObj();
        this.miniPlaylist.container.appendChild(this);
        return;
    };
}

function modifyImgElementPrototype() {
    HTMLImageElement.prototype.hide = function() {
        videoArr.remove(this);
        this.setAttribute("hidden", "hidden");
    };
    HTMLImageElement.prototype.show = function() {
        if (videoArr.indexOf(this) < 0) {
            // if it's not in the videoArr already, put it there.
            videoArr.push(this);
        }
        playerStarted(this.miniPlaylist.track);
        this.removeAttribute("hidden");
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
        this.miniPlaylist.playNext();
        this.hide();
        return;
    };
    HTMLImageElement.prototype.spawn = function(miniPlaylist, mediaObj) {
        this.miniPlaylist = miniPlaylist;
        this.mediaObj = mediaObj;
        this.setAttribute("class", "media");
        this.loadMediaObj();
        this.miniPlaylist.container.appendChild(this);
        return;
    }
    return;
}

function modifyIFrameElementPrototype() {
    HTMLIFrameElement.prototype.hide = function() {
        videoArr.remove(this);
        this.setAttribute("hidden", "hidden");
    };
    HTMLIFrameElement.prototype.show = function() {
        if (videoArr.indexOf(this) < 0) {
            // if it's not in the videoArr already, put it there.
            videoArr.push(this);
        }
        playerStarted(this.miniPlaylist.track);
        this.removeAttribute("hidden");
    };
    HTMLIFrameElement.prototype.loadMediaObj = function() {
        this.src = this.mediaObj.src;
    };
    HTMLIFrameElement.prototype.destroy = function() {
        this.parentNode.removeChild(this);
    };
    HTMLIFrameElement.prototype.setVolume = function(volumeLevel) {
        // should do nothing
        return;
    };
    HTMLIFrameElement.prototype.seekToEnd = function(volumeLevel) {
        this.miniPlaylist.playNext();
        this.hide();
        return;
    };
    HTMLIFrameElement.prototype.spawn = function(miniPlaylist, mediaObj) {
        this.miniPlaylist = miniPlaylist;
        this.mediaObj = mediaObj;
        this.setAttribute("class", "media");
        this.loadMediaObj();
        this.miniPlaylist.container.appendChild(this);
        return;
    }
    return;
}

function modifyYouTubePlayerPrototype() {
    // adjust YT.Player prototype for easier management
    YT.Player.prototype.isReady = false;
    YT.Player.prototype.startTime = 0.0;
    YT.Player.prototype.endTime = false;
    YT.Player.prototype.setReady = function() {
        this.setVolume(volume * trackVolume);
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
        var iframe = this.getIframe();
        iframe.parentNode.removeChild(iframe);
    };
    YT.Player.prototype.loadMediaObj = function() {
        if (this.isReady) {
            // if the player is ready
            params = {
                "videoId": this.mediaObj.src,
            }
            var start = this.mediaObj.start;
            var end = this.mediaObj.end;
            if (start.length > 0) {
                params.startSeconds = parseFloat(start);
                this.startTime = start;
            } else {
                params.startSeconds = 0;
            }
            if (end.length > 0) {
                params.endSeconds = parseFloat(end);
                this.endTime = end;
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
    YT.Player.prototype.containerId = null;
    YT.Player.prototype.soloLoop = false;
}

function modifySoundCloudPlayerPrototype(widget) {
    // adjust SoundCloud widget prototype for easier management
    widget.isReady = false;
    widget.startTime = 0.0;
    widget.endTime = false;
    widget.didEnd = false;
    // player should hold the mediaObj used to create it, so it can be
    // referenced later
    widget.mediaObj = false;
    widget.setReady = function() {
        this.setVolume(volume * trackVolume);
        if (this.mediaObj.muted) {
            this.mute();
        }
        this.isReady = true;
        if (this.mediaObj) {
            this.loadMediaObj();
        }
        return;
    };
    widget.gotDuration = function(value) {
        this.duration = value / 1000;
        this.endTime = this.endTime || this.duration;
    }
    widget.onReady = function() {
        widget.getDuration(widget.gotDuration);
        widget.setReady();
    };
    widget.bind(SC.Widget.Events.READY, widget.onReady);
    widget.onError = function() {
        var srcTrack = this.miniPlaylist.track;
        // remove all traces of the player
        this.miniPlaylist.playNext();
        submessage = "soundcloud player had an error";
        reportError(this.miniPlaylist.track, submessage);
        this.miniPlaylist.removePlayer(this);
        return;
    };
    widget.bind(SC.Widget.Events.ERROR, widget.onError);
    widget.hasEnded = function() {
        if (this.didEnd == false){
            this.didEnd = true;
            if (!this.soloLoop){
                this.hide();
                this.miniPlaylist.playNext();
            }
        }
    };
    widget.onPlayProgress = function(event) {
        this.currentTime = event.currentPosition / 1000;
        if (this.currentTime < this.startTime) {
            this.seekToStart();
        } else if (this.currentTime >= this.endTime){
            if (this.soloLoop){
                // when this is the only player in the playlist and the
                // playlist should loop, this should loop on it's own.
                this.currentTime = this.startTime;
                this.play();
            } else {
                this.pause();
                this.hasEnded();
                this.currentTime = this.startTime;
            }
        }
        return;
    };
    widget.bind(SC.Widget.Events.PLAY_PROGRESS, widget.onPlayProgress);
    widget.onPause = function(event) {
        this.hasEnded();
        return;
    };
    widget.bind(SC.Widget.Events.PAUSE, widget.onPause);
    widget.onFinish = function(event) {
        this.hasEnded();
        return;
    };
    widget.bind(SC.Widget.Events.FINISH, widget.onFinish);
    widget.seekToStart = function() {
        widget.seekTo(this.startTime * 1000); // takes milliseconds
        return;
    };
    widget.seekToEnd = function() {
        this.endTime = this.endTime;
        this.seekTo(this.endTime * 1000); // takes milliseconds
        return;
    };
    widget.getIframe = function() {
        return document.getElementById(this.containerId);
    };
    widget.destroy = function() {
        var iframe = this.getIframe();
        iframe.parentNode.removeChild(iframe);
        return;
    };
    widget.loadMediaObj = function() {
        if (this.isReady) {
            // if the player is ready
            var start = this.mediaObj.start;
            var end = this.mediaObj.end;
            if (start.length > 0) {
                this.seekTo(parseFloat(start));
                this.startTime = start;
            }
            if (end.length > 0) {
                this.endTime = end;
            }
            this.play();
        }
        return;
    };
    widget.hide = function() {
        this.getIframe().setAttribute("hidden", "hidden");
    };
    widget.show = function() {
        //should never show
        return;
    };
    widget.containerId = null;
    widget.soloLoop = false;
}

var VimeoPlayer = function(miniPlaylist, mediaObj) {
    //constructor for VimeoPlayer object class
    this.miniPlaylist = miniPlaylist;
    this.mediaObj = mediaObj;
    this.playerId = "vimeoplayer-" + Math.floor((Math.random() * 100000) + 1).toString();
    this.iframe = null;
    this.isReady = false;
    this.startTime = 0;
    this.endTime = false;
    this.didEnd = false;
    this.soloLoop = false;
    this.duration = null;
};

function modifyVimeoPlayerPrototype() {
    VimeoPlayer.prototype.spawn = function() {
        // create a <iframe> for the Vimeo player
        this.iframe = document.createElement("iframe");
        var player = this;
        this.iframe.getPlayer = function() {
            return player;
        }

        // use a unique ID so multiple players can be spawned and referenced
        this.iframe.setAttribute("id", this.playerId);
        this.iframe.setAttribute("class", "media");

        // hide the vimeo trackbar
        this.iframe.style.height = "200%";
        this.iframe.style.overflow = "hidden";
        this.iframe.style.transform = "translate(0, -25%)";

        // hide the player at first so it doesn't block anything before it
        // starts actually playing
        this.iframe.style.visibility = "hidden";
        // get trackID of song URL in order to embed the widget properly
        this.iframe.src = "http://player.vimeo.com/video/" +
            this.mediaObj.src + "?player_id=" + this.playerId +
            "&api=1&badge=0&byline=0&portrait=0&title=0&loop1";
        this.miniPlaylist.container.appendChild(this.iframe);
    };
    VimeoPlayer.prototype.onReady = function(event) {
        this.post("addEventListener", "playProgress");
        this.post("addEventListener", "play");
        this.post("addEventListener", "pause");
        this.post("addEventListener", "finish");
        this.storeDuration();
        this.setReady();
        
    };
    VimeoPlayer.prototype.setReady = function() {
        this.setVolume(volume * trackVolume);
        if (this.mediaObj.muted) {
            this.mute();
        }
        this.isReady = true;
        if (this.mediaObj) {
            this.loadMediaObj();
        }
        return;
    };
    VimeoPlayer.prototype.post = function(action, value) {
        var data = {
          method: action
        };
        if (value) {
            data.value = value;
        }
        var message = JSON.stringify(data);
        if (this.iframe.contentWindow) {
            // player may already be destroyed
            this.iframe.contentWindow.postMessage(message, "*");
        }
    };
    VimeoPlayer.prototype.storeDuration = function(level) {
        this.post("getDuration");
    };
    VimeoPlayer.prototype.setVolume = function(level) {
        this.post("setVolume", (level/100.0));
    };
    VimeoPlayer.prototype.mute = function() {
        this.post("setVolume", 0.0001);
    };
    VimeoPlayer.prototype.loadMediaObj = function() {
        if (this.isReady) {
            // if the player is ready 
            var start = this.mediaObj.start;
            var end = this.mediaObj.end;
            if (start.length > 0) {
                this.seekTo(parseFloat(start));
                this.startTime = start;
            }
            if (end.length > 0) {
                this.endTime = end;
            } else if (this.duration) {
                this.endTime = this.duration;
            }
            this.play();
        }
        return;
    };
    VimeoPlayer.prototype.seekTo = function(time) {
        this.post("seekTo", time);
    };
    VimeoPlayer.prototype.play = function() {
        this.post("play");
    };
    VimeoPlayer.prototype.pause = function() {
        this.post("pause");
    };
    VimeoPlayer.prototype.hide = function() {
        videoArr.remove(this);
        this.iframe.style.visibility = "hidden";
    };
    VimeoPlayer.prototype.show = function() {
        if (videoArr.indexOf(this) < 0) {
            // if it's not in the videoArr already, put it there.
            videoArr.push(this);
        }
        playerStarted(this.miniPlaylist.track);
        this.iframe.style.visibility = "visible";
    };
    VimeoPlayer.prototype.seekToStart = function() {
        if (this.startTime == 0) {
            this.startTime = 0.001; // vimeo requires a positive float
        }
        this.seekTo(this.startTime);
    };
    VimeoPlayer.prototype.seekToEnd = function() {
        this.endTime = this.endTime || this.duration;
        this.seekTo(this.endTime);
    };
    VimeoPlayer.prototype.onPlayProgress = function(message) {
        var currentTime = message.seconds;
        if (currentTime < this.startTime) {
            this.seekToStart();
        } else if (currentTime >= this.endTime){
            if (this.soloLoop){
                // when this is the only player in the playlist and the
                // playlist should loop, this should loop on it's own.
                this.seekToStart();
                this.play();
            } else if (!this.soloLoop) {
                this.hasEnded();
                this.seekTo(this.startTime);
            }
        }
        return;
    };
    VimeoPlayer.prototype.onPlay = function(event) {
        this.show();
        this.didEnd = false;
    };
    VimeoPlayer.prototype.onPause = function(event) {
        this.hasEnded();
        return;
    };
    VimeoPlayer.prototype.onFinish = function(event) {
        if (this.soloLoop){
            this.seekToStart();
            this.play();
            return;
        }
        this.hasEnded();
        return;
    };
    VimeoPlayer.prototype.hasEnded = function() {
        if (this.didEnd == false){
            this.didEnd = true;
            if (!this.soloLoop){
                this.hide();
                this.pause();
                this.seekToStart();
                this.miniPlaylist.playNext();
            }
        }
    };
    VimeoPlayer.prototype.destroy = function() {
        this.iframe.parentNode.removeChild(this.iframe);
        return;
    };
    VimeoPlayer.prototype.onDurationChange = function() {
        this.endTime = this.endTime || this.duration;
        return;
    }
}

function receiveMessage(event) {
    if ((/^https?:\/\/player.vimeo.com/).test(event.origin)) {
        var message = JSON.parse(event.data);
        vimeoPlayerMessageHandler(message);

    } else if (event.origin !== "http://localhost:8181" &&
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
    var containerId = Math.floor((Math.random() * 100000) + 1).toString();
    playerDiv.setAttribute("id", containerId);
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
            "onReady": onYTPlayerReady,
            "onStateChange": onYTPlayerStateChange,
            "onError": onYTPlayerError,
        },
    }
    var ytPlayer = new YT.Player(containerId, playerParams);
    ytPlayer.containerId = containerId;
    ytPlayer.miniPlaylist = miniPlaylist;
    ytPlayer.mediaObj = mediaObj;
    ytPlayer.loadMediaObj();

    return ytPlayer;
}

function spawnSoundCloudPlayer(miniPlaylist, mediaObj) {
    // create a <div> for the YouTube player to replace
    var playerIFrame = document.createElement("iframe");

    // use a unique ID so multiple players can be spawned and referenced
    var containerId = Math.floor((Math.random() * 100000) + 1).toString();
    playerIFrame.setAttribute("id", containerId);
    playerIFrame.setAttribute("class", "media");
    // hide it at first so it doesn't block anything before it starts actually
    // playing
    playerIFrame.setAttribute("hidden", "hidden");
    // get trackID of song URL in order to embed the widget properly
    resolvedURL = resolveSoundCloudURL(mediaObj.src);
    playerIFrame.src = "https://w.soundcloud.com/player/?url=" + resolvedURL;
    miniPlaylist.container.appendChild(playerIFrame);
    var scPlayer = SC.Widget(containerId);
    modifySoundCloudPlayerPrototype(scPlayer)
    scPlayer.containerId = containerId;
    scPlayer.miniPlaylist = miniPlaylist;
    scPlayer.mediaObj = mediaObj;
    scPlayer.loadMediaObj();
    return scPlayer;
}

function resolveSoundCloudURL(songURL) {
    var xmlHttp = new XMLHttpRequest();
    var resolveURL = "http://api.soundcloud.com/resolve?url=";
    resolveURL += songURL;
    resolveURL += "&client_id=" + soundcloudClientId;
    xmlHttp.open("GET", resolveURL, false); // synchronous
    xmlHttp.send(null);
    response = JSON.parse(xmlHttp.responseText);
    return response.uri;
}

function spawnVimeoPlayer(miniPlaylist, mediaObj) {
    var vimeoplayer = new VimeoPlayer(miniPlaylist, mediaObj);

    vimeoplayer.spawn();
    return vimeoplayer;
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
                event.target.miniPlaylist.playNext();
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

function vimeoPlayerMessageHandler(message) {
    var playerIframe = document.getElementById(message.player_id);
    if (!playerIframe){
        //player may already be destroyed
        return;
    }
    var player = playerIframe.getPlayer();
    if (message.event) {
        switch(message.event) {
            case "ready":
                player.onReady();
                break;
            case "play":
                player.onPlay(message.data);
                break;
            case "playProgress":
                player.onPlayProgress(message.data);
                break;
            case "pause":
                player.onPause(message.data);
                break;
            case "finish":
                player.onFinish(message.data);
                break;
        }
    } else if (message.method) {
        switch (message.method) {
            case "getDuration":
                player.duration = message.value;
                player.onDurationChange();
                break;
        }
    }
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
    if (newVolume !== undefined && newVolume !== null){
        volume = newVolume;
    }
    for (mp of miniPlaylistArr) {
        mp.setVolume(volume * trackVolume);
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
    while (miniPlaylistArr.length > 0){
        miniPlaylistArr[0].cleanup();
    }
    // it's safer to clear the arrays after body is bodied
    videoArr = [];
    return;
}

function skip() {
    if (miniPlaylistArr.length > 0) {
        miniPlaylistArr[0].skip();
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
    modifyIFrameElementPrototype();

    modifyYouTubePlayerPrototype();
    modifyVimeoPlayerPrototype();

    return;
}
