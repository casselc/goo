/* Compatibility facade for the shared passive resource observation renderer. */
(function(global){const api=typeof module!=='undefined'?require('./resource-native-view'):global.GooResourceNativeView;if(typeof module!=='undefined')module.exports=api;else global.GooHostDashboard=api;})(typeof window==='undefined'?globalThis:window);
