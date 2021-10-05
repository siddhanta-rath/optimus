"use strict";(self.webpackChunkoptimus=self.webpackChunkoptimus||[]).push([[152],{3905:function(e,t,n){n.d(t,{Zo:function(){return c},kt:function(){return m}});var r=n(7294);function i(e,t,n){return t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function a(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),n.push.apply(n,r)}return n}function o(e){for(var t=1;t<arguments.length;t++){var n=null!=arguments[t]?arguments[t]:{};t%2?a(Object(n),!0).forEach((function(t){i(e,t,n[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):a(Object(n)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(n,t))}))}return e}function l(e,t){if(null==e)return{};var n,r,i=function(e,t){if(null==e)return{};var n,r,i={},a=Object.keys(e);for(r=0;r<a.length;r++)n=a[r],t.indexOf(n)>=0||(i[n]=e[n]);return i}(e,t);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);for(r=0;r<a.length;r++)n=a[r],t.indexOf(n)>=0||Object.prototype.propertyIsEnumerable.call(e,n)&&(i[n]=e[n])}return i}var s=r.createContext({}),u=function(e){var t=r.useContext(s),n=t;return e&&(n="function"==typeof e?e(t):o(o({},t),e)),n},c=function(e){var t=u(e.components);return r.createElement(s.Provider,{value:t},e.children)},p={inlineCode:"code",wrapper:function(e){var t=e.children;return r.createElement(r.Fragment,{},t)}},d=r.forwardRef((function(e,t){var n=e.components,i=e.mdxType,a=e.originalType,s=e.parentName,c=l(e,["components","mdxType","originalType","parentName"]),d=u(n),m=i,f=d["".concat(s,".").concat(m)]||d[m]||p[m]||a;return n?r.createElement(f,o(o({ref:t},c),{},{components:n})):r.createElement(f,o({ref:t},c))}));function m(e,t){var n=arguments,i=t&&t.mdxType;if("string"==typeof e||i){var a=n.length,o=new Array(a);o[0]=d;var l={};for(var s in t)hasOwnProperty.call(t,s)&&(l[s]=t[s]);l.originalType=e,l.mdxType="string"==typeof e?e:i,o[1]=l;for(var u=2;u<a;u++)o[u]=n[u];return r.createElement.apply(null,o)}return r.createElement.apply(null,n)}d.displayName="MDXCreateElement"},681:function(e,t,n){n.r(t),n.d(t,{frontMatter:function(){return l},contentTitle:function(){return s},metadata:function(){return u},toc:function(){return c},default:function(){return d}});var r=n(7462),i=n(3366),a=(n(7294),n(3905)),o=["components"],l={},s="Installation",u={unversionedId:"getting-started/installation",id:"getting-started/installation",isDocsHomePage:!1,title:"Installation",description:"Installing Optimus on any system is straight forward. We provide pre-built binaries,",source:"@site/docs/getting-started/installation.md",sourceDirName:"getting-started",slug:"/getting-started/installation",permalink:"/optimus/docs/getting-started/installation",editUrl:"https://github.com/odpf/optimus/edit/master/docs/docs/getting-started/installation.md",tags:[],version:"current",lastUpdatedBy:"Arinda Arif",lastUpdatedAt:1633417061,formattedLastUpdatedAt:"10/5/2021",frontMatter:{},sidebar:"docsSidebar",previous:{title:"Introduction",permalink:"/optimus/docs/introduction"},next:{title:"Configurations",permalink:"/optimus/docs/getting-started/configuration"}},c=[{value:"MacOS",id:"macos",children:[]},{value:"Download Binaries",id:"download-binaries",children:[]},{value:"Compiling from source",id:"compiling-from-source",children:[{value:"Prerequisites",id:"prerequisites",children:[]},{value:"Build",id:"build",children:[]}]}],p={toc:c};function d(e){var t=e.components,n=(0,i.Z)(e,o);return(0,a.kt)("wrapper",(0,r.Z)({},p,n,{components:t,mdxType:"MDXLayout"}),(0,a.kt)("h1",{id:"installation"},"Installation"),(0,a.kt)("p",null,"Installing Optimus on any system is straight forward. We provide pre-built ",(0,a.kt)("a",{parentName:"p",href:"https://github.com/odpf/optimus/releases"},"binaries"),",\nDocker Images and support package managers."),(0,a.kt)("h2",{id:"macos"},"MacOS"),(0,a.kt)("p",null,"You can install Optimus using homebrew on macOS:"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-shell"},"brew install odpf/taps/optimus\noptimus version\n")),(0,a.kt)("h2",{id:"download-binaries"},"Download Binaries"),(0,a.kt)("p",null,"The client and server binaries are downloadable at the releases tab. There is\ncurrently no installer available. You have to add the Optimus binary to the PATH\nenvironment variable yourself or put the binary in a location that is already\nin your $PATH (e.g. /usr/local/bin, ...)."),(0,a.kt)("p",null,"Once installed, you should be able to run:"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-shell"},"optimus version\n")),(0,a.kt)("h2",{id:"compiling-from-source"},"Compiling from source"),(0,a.kt)("h3",{id:"prerequisites"},"Prerequisites"),(0,a.kt)("p",null,"Optimus requires the following dependencies:"),(0,a.kt)("ul",null,(0,a.kt)("li",{parentName:"ul"},"Golang (version 1.16 or above)"),(0,a.kt)("li",{parentName:"ul"},"Git")),(0,a.kt)("h3",{id:"build"},"Build"),(0,a.kt)("p",null,"Run the following commands to compile ",(0,a.kt)("inlineCode",{parentName:"p"},"optimus")," from source"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-shell"},"git clone git@github.com:odpf/optimus.git\ncd optimus\nmake build\n")),(0,a.kt)("p",null,"Use the following command to test"),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-shell"},"./optimus version\n")),(0,a.kt)("p",null,"Optimus service can be started with the following command although there are few required\n",(0,a.kt)("a",{parentName:"p",href:"/optimus/docs/getting-started/configuration"},"configurations")," for it to start."),(0,a.kt)("pre",null,(0,a.kt)("code",{parentName:"pre",className:"language-shell"},"./optimus serve\n")))}d.isMDXComponent=!0}}]);