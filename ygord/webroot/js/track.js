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
        // does nothing to prevent audio player becoming visible
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
        var type = this.nodeName.toLowerCase();
        switch (this.error.code) {
            case this.error.MEDIA_ERR_ABORTED:
                reportError(type + " file playback has been aborted");
                break;
            case this.error.MEDIA_ERR_NETWORK:
                reportError(type + " file download halted due to network error");
                break;
            case this.error.MEDIA_ERR_DECODE:
                reportError(type + " file could not be decoded");
                break;
            case this.error.MEDIA_ERR_SRC_NOT_SUPPORTED:
                if (this.networkState == HTMLMediaElement.NETWORK_NO_SOURCE) {
                    reportError(type + " file could not be found");
                } else {
                    reportError(type + " file format is not supported");
                }
                break;
            default:
                reportError("unknown " + type + "error: " + this.error.MEDIA_ERR_SRC_NOT_SUPPORTED)
                break;
        }
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
    HTMLMediaElement.prototype.loadMedia = function() {
        var e = this.media.end;
        if (e.length > 0) {
            this.endTime = e;
        }
        this.muted = this.media.muted;
        this.src = this.media.src;
    };
    HTMLMediaElement.prototype.seekToEnd = function() {
        this.currentTime = this.endTime;
    };
    HTMLMediaElement.prototype.destroy = function() {
        this.parentNode.removeChild(this);
    };
    HTMLMediaElement.prototype.spawn = function(media) {
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
        this.media = media;
        this.soloLoop = media.loop;
        this.endTime = false;
        this.didEnd = false;
        this.setVolume(volume * trackVolume);
        this.loadMedia();
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
    HTMLImageElement.prototype.loadMedia = function() {
        this.src = this.media.src;
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
    HTMLImageElement.prototype.spawn = function(media) {
        this.media = media;
        this.setAttribute("class", "media");
        this.loadMedia();
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
    HTMLIFrameElement.prototype.loadMedia = function() {
        this.src = this.media.src;
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
    HTMLIFrameElement.prototype.spawn = function(media) {
        this.media = media;
        this.setAttribute("class", "media");
        this.loadMedia();
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
        if (!this.media.muted) {
            this.setVolume(volume * trackVolume);
        }
        var iframe = this.getIframe();
        iframe.player = this;
        if (this.media.muted) {
            this.mute();
        }
        this.isReady = true;
        // player may have been given a media before the player was ready,
        // so it should now load the desired video with the specified
        // parameters
        if (this.media) {
            this.loadMedia();
        }
    };
    // player should hold the media used to create it, so it can be
    // referenced later
    YT.Player.prototype.media = false;
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
        iframe.destroy();
    };
    YT.Player.prototype.loadMedia = function() {
        if (this.isReady) {
            // if the player is ready
            var params = {
                "videoId": this.media.src
            }
            var end = this.media.end;
            if (end.length > 0) {
                params.endSeconds = parseFloat(end);
                this.endTime = end;
            }
            this.soloLoop = this.media.loop;
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

function spawnYouTubePlayer(media) {
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
    var playerParams = {
        height: "100%",
        width: "100%",
        playerVars :{
            "controls": 0,
            "showinfo": 0,
            "rel": 0,
            "modestbranding": 1,
            "iv_load_policy": 3,
            "enablejsapi": 1,
            "origin": "https://ygor.truveris.com"
        },
        events: {
            "onReady": onYTPlayerReady,
            "onStateChange": onYTPlayerStateChange,
            "onError": onYTPlayerError
        }
    }
    var ytPlayer = new YT.Player(containerId, playerParams);
    ytPlayer.containerId = containerId;
    ytPlayer.media = media;
    ytPlayer.loadMedia();

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
    switch(event.data){
        case 2:
            reportError("invalid youtube video parameter");
            break;
        case 5:
            reportError("youtube video doesn't work with html5");
            break;
        case 100:
            reportError("no such youtube video");
            break;
        case 101:
            reportError("can't embed this youtube video");
            break;
        case 150:
            reportError("can't embed this youtube video");
            break;
        default:
            reportError("unrecognized error code: " + event.data);
            break;
    }
    // remove all traces of the player
    playerEnded();
    this.destroy();
    return;
}

/* ------------------------------ END YOUTUBE ------------------------------ */

/* --------------------------------- VIMEO --------------------------------- */

var VimeoPlayer = function(media) {
    //constructor for VimeoPlayer object class
    this.media = media;
    this.playerId = "vimeoplayer-" + Math.floor((Math.random() * 100000) + 1).toString();
    this.iframe = null;
    this.isReady = false;
    // vimeo requires a positive float
    this.startTime = 0.001;
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
        this.iframe.player = this;

        // use a unique ID so multiple players can be spawned and referenced
        this.iframe.setAttribute("id", this.playerId);
        this.iframe.setAttribute("class", "media");

        // hide the vimeo trackbar
        this.iframe.style.height = "200%";
        this.iframe.style.overflow = "hidden";
        this.iframe.style.transform = "translate(0, -25%)";

        // hide the player at first so it doesn't block anything before it
        // starts actually playing
        this.iframe.show();
        // get trackID of song URL in order to embed the widget properly
        var looping = "";
        if (this.media.loop) {
            looping = "&loop=1";
        }
        this.iframe.src = "https://player.vimeo.com/video/" +
            this.media.src + "?player_id=" + this.playerId +
            "&api=1&badge=0&byline=0&portrait=0&title=0" + looping;
        document.body.appendChild(this.iframe);
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
        if (this.media.muted) {
            this.mute();
        }
        this.isReady = true;
        if (this.media) {
            this.loadMedia();
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
        if (this.iframe.contentWindow) {
            // player may already be destroyed
            this.iframe.contentWindow.postMessage(data, "*");
        }
    };
    VimeoPlayer.prototype.storeDuration = function() {
        this.post("getDuration");
    };
    VimeoPlayer.prototype.setVolume = function(level) {
        if (level == 0) {
            // must be positive float
            this.mute();
        } else {
            this.post("setVolume", (level/100.0));
        }
    };
    VimeoPlayer.prototype.mute = function() {
        this.post("setVolume", 0.0001);
    };
    VimeoPlayer.prototype.loadMedia = function() {
        if (this.isReady) {

            // if the player is ready
            var end = this.media.end;
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
        this.iframe.hide();
    };
    VimeoPlayer.prototype.show = function() {
        this.iframe.show();
        playerStarted();
    };
    VimeoPlayer.prototype.seekToStart = function() {
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
                this.seekToStart();
                this.play();
            } else if (!this.soloLoop) {
                this.hasEnded();
                this.seekToStart();
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
                playerEnded();
                this.destroy();
            }
        }
    };
    VimeoPlayer.prototype.destroy = function() {
        this.iframe.destroy();
        delete this;
    };
    VimeoPlayer.prototype.onDurationChange = function() {
        this.endTime = this.endTime || this.duration;
        return;
    }
}

function spawnVimeoPlayer(media) {
    var vimeoplayer = new VimeoPlayer(media);

    vimeoplayer.spawn();
    return vimeoplayer;
}

function vimeoPlayerMessageHandler(message) {
    var playerIframe = document.getElementById(message.player_id);
    if (!playerIframe){
        //player may already be destroyed
        return;
    }
    var player = playerIframe.player;
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

/* ------------------------------- END VIMEO ------------------------------- */

/* ------------------------------- SOUNDCLOUD ------------------------------ */

var SoundCloudPlayer = function(media) {
    //constructor for SoundCloudPlayer object class
    this.media = media;
    this.playerId = "soundcloudplayer-" + Math.floor((Math.random() * 100000) + 1).toString();
    this.iframe = null;
    this.isReady = false;
    // vimeo requires a positive float
    this.startTime = 0.0;
    this.endTime = false;
    this.didEnd = false;
    this.duration = null;
    // methods
    this.spawn = function() {
        // create a <div> for the YouTube player to replace
        this.iframe = document.createElement("iframe");

        // use a unique ID so multiple players can be spawned and referenced
        this.iframe.setAttribute("id", this.playerId);
        this.iframe.setAttribute("class", "media");
        this.iframe.hide();
        // hide it at first so it doesn't block anything before it starts actually
        // playing
        //this.iframe.setAttribute("hidden", "hidden");
        this.iframe.src = "https://w.soundcloud.com/player/?url=" + this.media.src;
        document.body.appendChild(this.iframe);
        var scPlayer = SC.Widget(this.playerId);
        this.iframe.player = this;
        this.player = scPlayer;
        modifySoundCloudPlayerPrototype(this.player)
        this.player.containerId = this.playerId;
        this.player.media = this.media;
        this.player.loadMedia();
    }
    this.setVolume = function(level) {
        // SoundCloud Widget requires float between 0 and 1
        this.player.setVolume(level * 0.01);
    }
};

function modifySoundCloudPlayerPrototype(widget) {
    // adjust SoundCloud widget prototype for easier management
    widget.isReady = false;
    widget.startTime = 0.0;
    widget.endTime = false;
    widget.didEnd = false;
    // player should hold the media used to create it, so it can be
    // referenced later
    widget.media = false;
    widget.getIframe = function() {
        return document.getElementById(this.containerId);
    };
    widget.setReady = function() {
        // SoundCloud player requires float between 0 and 1 for volume
        this.setVolume(volume * trackVolume * 0.01);
        this.isReady = true;
        if (this.media) {
            this.loadMedia();
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
        // remove all traces of the player
        reportError("soundcloud player had an error");
        playerEnded();
        this.destroy();
        return;
    };
    widget.bind(SC.Widget.Events.ERROR, widget.onError);
    widget.hasEnded = function() {
        if (this.didEnd == false){
            this.didEnd = true;
            playerEnded();
            this.destroy();
        }
    };
    widget.onPlayProgress = function(event) {
        this.currentTime = event.currentPosition / 1000;
        if (this.currentTime < this.startTime) {
            this.seekToStart();
        } else if (this.currentTime >= this.endTime){
            this.pause();
            this.hasEnded();
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
    widget.destroy = function() {
        var iframe = this.getIframe();
        iframe.destroy();
        return;
    };
    widget.loadMedia = function() {
        if (this.isReady) {
            // if the player is ready
            var end = this.media.end;
            if (end.length > 0) {
                this.endTime = end;
            }
            this.play();
        }
        return;
    };
    widget.hide = function() {
        this.getIframe().hide();
    };
    widget.show = function() {
        //should never show
        return;
    };
    widget.containerId = null;
    widget.soloLoop = false;
}

function spawnSoundCloudPlayer(media) {
    var scPlayer = new SoundCloudPlayer(media);

    scPlayer.spawn();
}

/* ----------------------------- END SOUNDCLOUD ---------------------------- */

function receiveMessage(event) {
    var media = event.data;

    var vimeoRe = /^https?:\/\/player\.vimeo\.com/
    if (vimeoRe.test(event.origin)) {
        vimeoPlayerMessageHandler(JSON.parse(media));
        return;
    }

    /* Ignore all messages that are not sent from the parent frame. */
    if (event.origin !== window.location.origin) {
        return;
    }

    spawnMedia(media);
}

function spawnMedia(media) {
    switch (media.format){
        case "video":
            spawnStandardPlayer(media);
            break;
        case "audio":
            spawnStandardPlayer(media);
            break;
        case "img":
            spawnStandardPlayer(media);
            break;
        case "web":
            spawnWeb(media);
            break;
        case "youtube":
            spawnYouTubePlayer(media);
            break;
        case "vimeo":
            spawnVimeoPlayer(media);
            break;
        case "soundcloud":
            spawnSoundCloudPlayer(media);
            break;
        default:
            reportError("unrecognized format: " + media.format)
    }
}

// spawns <video>, <audio>, or <img>
function spawnStandardPlayer(media) {
    var player = document.createElement(media.format);
    player.spawn(media);
    return player;
}

// spawns iframe to show webpage
function spawnWeb(media) {
    var web = document.createElement("iframe");
    web.spawn(media);
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
    var count = 0;
    for(el of getAllPlayers()) {
        if(el.getAttribute("opacity") != 0) {
            count++;
        }
    }
    return count;
}

function getAllPlayers() {
    var collection = document.querySelectorAll("body > *");
    var elementArr = [];
    for(i = 0; i < collection.length; i++){
        elementArr.push(collection.item(i));
    }
    return elementArr;
}

function reportError(submessage) {
    var submessage = submessage || "";
    sendMessage("ERRORED", submessage);
    return;
}

function sendMessage(state, submessage) {
    var message = {
        playerState: state,
        submessage: submessage
    }
    parent.postMessage(message, "*");
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
    modifyVimeoPlayerPrototype();

    return;
}
