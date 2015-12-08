var ygorMinion = angular.module("ygorMinion", [
	"ngRoute",
	"ngSanitize",
	"ygorMinionControllers"
]);


ygorMinion.config(["$routeProvider",
    function($routeProvider) {
        $routeProvider.
            when("/menu", {
                templateUrl: "partials/menu.html"
            }).
            when("/alias/list", {
                templateUrl: "partials/alias-list.html",
                controller: "AliasListController"
            }).
            when("/channel/list", {
                templateUrl: "partials/channel-list.html",
                controller: "ChannelListController"
            }).
            when("/client/list", {
                templateUrl: "partials/client-list.html",
                controller: "ClientListController"
            }).
            when("/channel/:channelID", {
                templateUrl: "partials/channel.html",
                controller: "ChannelController"
            }).
            otherwise({
                redirectTo: "/menu"
            });
    }
]);
