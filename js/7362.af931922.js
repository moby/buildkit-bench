"use strict";(self["webpackChunkbuildkit_bench_website"]=self["webpackChunkbuildkit_bench_website"]||[]).push([[7362],{7362:(e,a,t)=>{t.r(a),t.d(a,{default:()=>v});var l=' <!DOCTYPE html> <html> <head> <meta charset="utf-8"> <title>BuildKit Benchmarks</title> <script src="https://go-echarts.github.io/go-echarts-assets/assets/echarts.min.js"><\/script> </head> <body> <style>.box{justify-content:center;display:flex;flex-wrap:wrap}</style> <div class="box"> <div class="container"> <div class="item" id="1520259285885722123" style="width:900px;height:500px"></div> </div><script>"use strict";let goecharts_1520259285885722123=echarts.init(document.getElementById("1520259285885722123"),"white",{renderer:"canvas"}),option_1520259285885722123={color:["#5470c6","#91cc75","#fac858","#ee6666","#73c0de","#3ba272","#fc8452","#9a60b4","#ea7ccc"],dataZoom:[{type:"slider"}],legend:{},series:[{name:"Time (s)",type:"boxplot",data:[{value:[59.688265957,59.711107559,59.738315370500004,59.833226915,59.92377225]},{value:[59.958715571,60.020116393500004,60.243263260999996,60.485672332,60.566335358]},{value:[83.922361649,83.93353286249999,84.358120336,84.993331547,85.215126498]},{value:[10.781433447,10.790286518999999,10.8437252125,10.9600237205,11.031736607]},{value:[22.852226544,23.1624914535,23.8099303945,25.217075845,26.287047264]},{value:[22.171879334,22.6181916365,23.2223924625,23.900086572,24.419892158]},{value:[22.183476031,22.265569711,22.370615159,23.392602066,24.391637205]},{value:[21.274564914,22.1289519205,23.1653084985,23.7229490625,24.098620055]},{value:[23.488962105,23.528189036,24.266377433,26.043002154,27.120665409]},{value:[22.098480239,22.217120409,23.124858095,24.5444459785,25.174936346]},{value:[5.960885559,5.9774311055,6.002758362,6.1065942815,6.201648491]},{value:[5.805084724,5.872445698,5.956375672,6.0131360745,6.053327477]},{value:[5.768137826,5.8487545615,5.958345298,6.019315478999999,6.051311659]},{value:[5.762364431,5.777920524,5.8014384705,6.190253814,6.571107304]},{value:[6.093317975,6.112151584999999,6.216687199,6.330500098,6.358610993]},{value:[5.856657248,5.858176624,5.978926475,6.1608563985,6.223555847]},{value:[6.008946669,6.0459464364999995,6.119055056,6.2346112675,6.314058627]},{value:[5.862925567,5.9464802599999995,6.041859882000001,6.058069096500001,6.062453382]},{value:[5.972450638,6.0374922845,6.1385825115,6.287962214,6.401293336]},{value:[5.779725838,5.903708418,6.1152371275,6.2966242975,6.390465338]},{value:[6.018837024,6.032621262999999,6.0589480285,6.20607201,6.340653465]},{value:[5.688687349,5.8336442680000005,5.997157726499999,6.238548199,6.461382132]},{value:[5.840519631,6.0912438945,6.350139260500001,6.4013463005,6.444382238]},{value:[5.843806979,5.9245789994999996,6.08906083,6.307185767,6.4416008940000005]},{value:[5.608688488,5.8703729835,6.138760782,6.254326421,6.363188757]}]}],title:{text:"Build breaker 128x",subtext:"BenchmarkBuild/BenchmarkBuildBreaker128"},toolbox:{show:!0,orient:"horizontal",right:"100",feature:{saveAsImage:{show:!0},dataZoom:{show:!0,icon:null,title:null},restore:{show:!0}}},tooltip:{},xAxis:[{data:["v0.9.0","v0.9.3","v0.10.0","v0.10.6","v0.11.0","v0.11.6","v0.12.0","v0.12.5","v0.13.0","v0.13.2","v0.14.0","v0.14.1","v0.15.0","v0.15.2","2024-09-04","2024-09-05","2024-09-06","2024-09-09","2024-09-10","v0.16.0","2024-09-11","2024-09-12","2024-09-13","2024-09-16","master"]}],yAxis:[{}]};goecharts_1520259285885722123.setOption(option_1520259285885722123)<\/script> <div class="container"> <div class="item" id="1818003733346867093" style="width:900px;height:500px"></div> </div><script>"use strict";let goecharts_1818003733346867093=echarts.init(document.getElementById("1818003733346867093"),"white",{renderer:"canvas"}),option_1818003733346867093={color:["#5470c6","#91cc75","#fac858","#ee6666","#73c0de","#3ba272","#fc8452","#9a60b4","#ea7ccc"],dataZoom:[{type:"slider"}],legend:{},series:[{name:"Time (s)",type:"boxplot",data:[{value:[6.930054306,6.930054306,6.933235937,6.947134522,6.947134522]},{value:[6.925294624,6.925294624,6.93253859,6.932905691,6.932905691]},{value:[9.714161837,9.714161837,9.735819659,9.782736821,9.782736821]},{value:[1.3406069409999999,1.3406069409999999,1.347037317,1.362446988,1.362446988]},{value:[1.873236428,1.873236428,1.977724123,2.001597697,2.001597697]},{value:[1.901985617,1.901985617,1.9170085559999999,1.969391892,1.969391892]},{value:[1.8213073020000001,1.8213073020000001,1.854998616,1.880463646,1.880463646]},{value:[1.8766399809999998,1.8766399809999998,1.901846295,1.915146271,1.915146271]},{value:[1.91192227,1.91192227,1.9612636490000002,1.976870975,1.976870975]},{value:[1.909291643,1.909291643,1.911084098,1.921372973,1.921372973]},{value:[.683017515,.683017515,.764032153,.776157419,.776157419]},{value:[.735267192,.735267192,.752934882,.778625463,.778625463]},{value:[.758711317,.758711317,.769438368,.789908052,.789908052]},{value:[.686928757,.686928757,.761131982,1.191720829,1.191720829]},{value:[.77821881,.77821881,.786337433,.80347681,.80347681]},{value:[.690806845,.690806845,.785994693,.807227561,.807227561]},{value:[.674148883,.674148883,.677723196,1.171723602,1.171723602]},{value:[.710160159,.710160159,.761462703,.783706668,.783706668]},{value:[.801430088,.801430088,.804797308,.819531757,.819531757]},{value:[.750542276,.750542276,.775528528,.790837145,.790837145]},{value:[.762676296,.762676296,.793310586,1.188839363,1.188839363]},{value:[.723463022,.723463022,.800782071,.817834868,.817834868]},{value:[.717430503,.717430503,.757185416,.828811301,.828811301]},{value:[.80163539,.80163539,.815427122,.834721512,.834721512]},{value:[.786527975,.786527975,.80171262,.801921405,.801921405]}]}],title:{text:"Build breaker 16x",subtext:"BenchmarkBuild/BenchmarkBuildBreaker16"},toolbox:{show:!0,orient:"horizontal",right:"100",feature:{saveAsImage:{show:!0},dataZoom:{show:!0,icon:null,title:null},restore:{show:!0}}},tooltip:{},xAxis:[{data:["v0.9.0","v0.9.3","v0.10.0","v0.10.6","v0.11.0","v0.11.6","v0.12.0","v0.12.5","v0.13.0","v0.13.2","v0.14.0","v0.14.1","v0.15.0","v0.15.2","2024-09-04","2024-09-05","2024-09-06","2024-09-09","2024-09-10","v0.16.0","2024-09-11","2024-09-12","2024-09-13","2024-09-16","master"]}],yAxis:[{}]};goecharts_1818003733346867093.setOption(option_1818003733346867093)<\/script> <div class="container"> <div class="item" id="7260912646843490284" style="width:900px;height:500px"></div> </div><script>"use strict";let goecharts_7260912646843490284=echarts.init(document.getElementById("7260912646843490284"),"white",{renderer:"canvas"}),option_7260912646843490284={color:["#5470c6","#91cc75","#fac858","#ee6666","#73c0de","#3ba272","#fc8452","#9a60b4","#ea7ccc"],dataZoom:[{type:"slider"}],legend:{},series:[{name:"Time (s)",type:"boxplot",data:[{value:[14.118909156,14.124212679,14.142296417499999,14.168793166,14.182509699]},{value:[14.264047093,14.285089236000001,14.3264239605,14.36577369,14.384830838]},{value:[20.024340995,20.0540482605,20.097884569999998,20.12151913,20.131024646]},{value:[2.719077495,2.7252202575,2.7315555424999998,2.7391626220000003,2.746577179]},{value:[4.072661575,4.118928028999999,4.2351256615,4.333113273,4.361169706]},{value:[4.098985017,4.190342874000001,4.357460773,4.446952539,4.460684263]},{value:[4.164724466,4.2032349045,4.2593390825,4.3256864,4.374439978]},{value:[4.086482549,4.2029924350000005,4.328416381,4.341323711499999,4.345316982]},{value:[4.536657791,4.5396333945,4.553311583,4.577969796,4.591925424]},{value:[4.082659062,4.263072277,4.487323824,4.5430121415,4.554862127]},{value:[1.387103818,1.3989041985,1.4288805514999998,1.4514342205,1.4558119170000001]},{value:[1.327153217,1.3345560695,1.3698441035,1.4821413465,1.5665534079999999]},{value:[1.329602132,1.3424595245000002,1.3623904665,1.385224894,1.400985772]},{value:[1.270115429,1.2872267995,1.3120164225000002,1.387736248,1.455777821]},{value:[1.421384419,1.4316231655,1.441984245,1.4901790314999999,1.538251485]},{value:[1.442345419,1.4585480815,1.476457056,1.5006207665,1.523078165]},{value:[1.380984056,1.439491206,1.503021946,1.5521650375,1.596284539]},{value:[1.441582587,1.4575030015000001,1.4756791644999998,1.511318922,1.544702931]},{value:[1.457904627,1.459560894,1.463729896,1.49366449,1.521086349]},{value:[1.42352528,1.4356100695,1.4610219359999999,1.4991911464999998,1.52403328]},{value:[1.4266196039999999,1.452691553,1.482179482,1.5084649685,1.531334475]},{value:[1.438557659,1.4469245675,1.4872109425,1.5347046734999998,1.550278938]},{value:[1.505399622,1.512881787,1.5226522775,1.536030905,1.547121207]},{value:[1.465808642,1.473179148,1.4852638425,1.5059579885,1.521937946]},{value:[1.457173189,1.4841169145,1.5137418655000001,1.5372604,1.558097709]}]}],title:{text:"Build breaker 32x",subtext:"BenchmarkBuild/BenchmarkBuildBreaker32"},toolbox:{show:!0,orient:"horizontal",right:"100",feature:{saveAsImage:{show:!0},dataZoom:{show:!0,icon:null,title:null},restore:{show:!0}}},tooltip:{},xAxis:[{data:["v0.9.0","v0.9.3","v0.10.0","v0.10.6","v0.11.0","v0.11.6","v0.12.0","v0.12.5","v0.13.0","v0.13.2","v0.14.0","v0.14.1","v0.15.0","v0.15.2","2024-09-04","2024-09-05","2024-09-06","2024-09-09","2024-09-10","v0.16.0","2024-09-11","2024-09-12","2024-09-13","2024-09-16","master"]}],yAxis:[{}]};goecharts_7260912646843490284.setOption(option_7260912646843490284)<\/script> <div class="container"> <div class="item" id="955529895840668285" style="width:900px;height:500px"></div> </div><script>"use strict";let goecharts_955529895840668285=echarts.init(document.getElementById("955529895840668285"),"white",{renderer:"canvas"}),option_955529895840668285={color:["#5470c6","#91cc75","#fac858","#ee6666","#73c0de","#3ba272","#fc8452","#9a60b4","#ea7ccc"],dataZoom:[{type:"slider"}],legend:{},series:[{name:"Time (s)",type:"boxplot",data:[{value:[28.445183688,28.678547213999998,28.9477610435,28.9963765385,29.00914173]},{value:[28.300260487,28.393577305,28.6644419585,29.022914139,29.203838484]},{value:[39.756085426,39.819967247,40.0243829865,40.289549093,40.414181281]},{value:[5.23416686,5.235146782,5.237836258,5.270304776,5.30106374]},{value:[7.827530236,7.861104549,8.722412499,10.211433795000001,10.872721454]},{value:[8.349190921,8.4660937625,8.7695641685,9.1940310265,9.43193032]},{value:[8.500530356,9.084864186,9.6882337895,9.73214775,9.757025937]},{value:[8.576301904,8.6779314585,9.0080161785,9.398194936,9.559918528]},{value:[9.31591755,9.458968723,9.857224197,10.2292803335,10.346132169]},{value:[8.526979925,8.929245284,9.478821801999999,10.307595016499999,10.989057072]},{value:[2.660578067,2.6809818080000003,2.7114830725,2.8780251865,3.034469777]},{value:[2.69653438,2.7104929925,2.733649078,2.9254232105,3.10799987]},{value:[2.68394479,2.695994103,2.712949378,2.756304225,2.7947531100000003]},{value:[2.6826948809999998,2.7204743359999997,2.8091604295000003,2.881211627,2.902356186]},{value:[2.665744068,2.6875065535,2.724822461,2.7842951485,2.828214414]},{value:[2.808401521,2.83583607,2.8982870245,3.038829315,3.1443552]},{value:[2.6215023410000002,2.6505820815,2.7971694955,3.0485587174999997,3.182440266]},{value:[2.513627218,2.6007358115,2.7112689405,2.787388945,2.840084414]},{value:[2.625255376,2.63165815,2.680623613,2.7622815215000003,2.801376741]},{value:[2.716304359,2.7690122795,2.8270773,2.8797114859999997,2.926988572]},{value:[2.697292456,2.7668467960000003,2.836747246,2.8913160935,2.945538831]},{value:[2.785492209,2.847216496,2.9121913370000003,2.9800970275000003,3.044752164]},{value:[2.684446355,2.7419024015,2.8344438524999997,2.9827565875,3.095983918]},{value:[2.904923986,2.9210883515,3.0007370355,3.0776313935,3.091041433]},{value:[2.760514291,2.793853426,2.83532022,2.8655921810000002,2.887736483]}]}],title:{text:"Build breaker 64x",subtext:"BenchmarkBuild/BenchmarkBuildBreaker64"},toolbox:{show:!0,orient:"horizontal",right:"100",feature:{saveAsImage:{show:!0},dataZoom:{show:!0,icon:null,title:null},restore:{show:!0}}},tooltip:{},xAxis:[{data:["v0.9.0","v0.9.3","v0.10.0","v0.10.6","v0.11.0","v0.11.6","v0.12.0","v0.12.5","v0.13.0","v0.13.2","v0.14.0","v0.14.1","v0.15.0","v0.15.2","2024-09-04","2024-09-05","2024-09-06","2024-09-09","2024-09-10","v0.16.0","2024-09-11","2024-09-12","2024-09-13","2024-09-16","master"]}],yAxis:[{}]};goecharts_955529895840668285.setOption(option_955529895840668285)<\/script> <div class="container"> <div class="item" id="315869857204764781" style="width:900px;height:500px"></div> </div><script>"use strict";let goecharts_315869857204764781=echarts.init(document.getElementById("315869857204764781"),"white",{renderer:"canvas"}),option_315869857204764781={color:["#5470c6","#91cc75","#fac858","#ee6666","#73c0de","#3ba272","#fc8452","#9a60b4","#ea7ccc"],dataZoom:[{type:"slider"}],legend:{},series:[{name:"Time (s)",type:"boxplot",data:[{value:[3.518146606,3.525353168,3.567468948,3.6034326035,3.604487041]},{value:[3.514353931,3.5284056479999997,3.568796896,3.604518228,3.613900029]},{value:[3.682068436,3.6987124089999996,3.7252217004999997,3.7375591175,3.740031216]},{value:[3.23877988,3.239220334,3.248507795,3.3216386475,3.385922493]},{value:[3.2804788,3.2854469010000003,3.291641538,3.298189027,3.30350998]},{value:[3.297953184,3.3005070185,3.308430877,3.319751524,3.325702147]},{value:[3.184485069,3.1973772714999997,3.215594123,3.236988185,3.253057598]},{value:[3.211552931,3.212762517,3.2229758349999997,3.3187082455,3.405436924]},{value:[2.9321483649999998,2.9348017769999997,2.940792469,2.9456561525,2.947182556]},{value:[2.913336056,2.9176698175,2.9325407745,2.9682475475,2.993417125]},{value:[2.8705758120000002,2.8816785325,2.8952430555,2.910887071,2.9240692839999998]},{value:[2.874385547,2.884059539,2.8993880020000002,2.907626821,2.910211169]},{value:[2.874727788,2.8753854115,2.882333009,2.8997266364999996,2.91083029]},{value:[2.903225215,2.908190647,2.929143526,2.9631074909999997,2.981084009]},{value:[2.848412453,2.8805342549999997,2.9280304824999996,2.963948216,2.984491524]},{value:[2.8831371690000003,2.8879655415,2.8952809659999996,2.8990762445,2.9003844709999997]},{value:[2.884464152,2.8907250744999997,2.9030787665,2.9406751480000004,2.9721787600000003]},{value:[2.900459244,2.913620357,2.9273158665,2.9538941029999997,2.979937943]},{value:[2.870083077,2.871174235,2.8838461530000004,2.9115351580000004,2.9276434030000003]},{value:[2.852675632,2.863209245,2.8803943115,2.8878397645,2.888633764]},{value:[2.871303314,2.8729983555,2.874844081,2.890675978,2.906357191]},{value:[2.827317375,2.8495063685,2.8830368075,2.9357205420000003,2.977062831]},{value:[2.859351199,2.8790314065,2.91631158,2.9381913145,2.942471083]},{value:[2.827934273,2.83352879,2.8395991030000003,2.8622454910000004,2.884416083]},{value:[2.833683158,2.8431189884999997,2.881256909,2.912446547,2.914934095]}]}],title:{text:"Build with substantial file transfer",subtext:"BenchmarkBuild/BenchmarkBuildFileTransfer"},toolbox:{show:!0,orient:"horizontal",right:"100",feature:{saveAsImage:{show:!0},dataZoom:{show:!0,icon:null,title:null},restore:{show:!0}}},tooltip:{},xAxis:[{data:["v0.9.0","v0.9.3","v0.10.0","v0.10.6","v0.11.0","v0.11.6","v0.12.0","v0.12.5","v0.13.0","v0.13.2","v0.14.0","v0.14.1","v0.15.0","v0.15.2","2024-09-04","2024-09-05","2024-09-06","2024-09-09","2024-09-10","v0.16.0","2024-09-11","2024-09-12","2024-09-13","2024-09-16","master"]}],yAxis:[{}]};goecharts_315869857204764781.setOption(option_315869857204764781)<\/script> <div class="container"> <div class="item" id="4211100011286834090" style="width:900px;height:500px"></div> </div><script>"use strict";let goecharts_4211100011286834090=echarts.init(document.getElementById("4211100011286834090"),"white",{renderer:"canvas"}),option_4211100011286834090={color:["#5470c6","#91cc75","#fac858","#ee6666","#73c0de","#3ba272","#fc8452","#9a60b4","#ea7ccc"],dataZoom:[{type:"slider"}],legend:{},series:[{name:"Time (s)",type:"boxplot",data:[{value:[.668136888,.674281369,.6833535855,.6959549095,.705628498]},{value:[.670472947,.670634069,.671713754,.6748315745,.677030832]},{value:[.873273208,.876097965,.8799210605000001,.8815060495,.8820927]},{value:[.348913855,.3493092825,.3503282745,.3523459745,.35374011]},{value:[.358212633,.368441727,.38214176099999997,.39485162,.404090539]},{value:[.363407233,.367948469,.37609876799999997,.38729496799999996,.394882105]},{value:[.371288168,.377617717,.387586955,.3944534435,.397680243]},{value:[.366822538,.3772104245,.39147504499999997,.395531514,.395711249]},{value:[.358185173,.36378847950000004,.3713622005,.37360201,.373871405]},{value:[.358626482,.3627276435,.3708343035,.375828896,.37681799]},{value:[.286973779,.2872005315,.288314104,.29396023950000005,.298719555]},{value:[.2792632,.28059571549999995,.2827265515,.28365553649999997,.283786201]},{value:[.278785731,.286016593,.29474768549999997,.30186356950000004,.307479223]},{value:[.288712959,.29222626549999997,.296368575,.2974267195,.297855861]},{value:[.276719689,.279957107,.284603964,.2885904995,.291167596]},{value:[.273744807,.279741417,.293445163,.3063776515,.311603004]},{value:[.263250642,.26719429949999995,.275437171,.28347175599999996,.287207127]},{value:[.270595329,.272780902,.2792998165,.2849300305,.286226903]},{value:[.271912855,.275107141,.28215311649999997,.287960872,.289916938]},{value:[.279317293,.28328079250000004,.290909094,.3018044905,.309035085]},{value:[.275497564,.277226086,.2807750545,.29004423300000004,.297492965]},{value:[.276560742,.2773227815,.283786237,.2921502575,.294812862]},{value:[.269670141,.27406388049999997,.2790138775,.280772636,.281975137]},{value:[.274123276,.278010529,.2853934775,.290017007,.291144841]},{value:[.266075241,.269583195,.277730218,.28391518800000004,.285461089]}]}],title:{text:"Multistage build",subtext:"BenchmarkBuild/BenchmarkBuildMultistage"},toolbox:{show:!0,orient:"horizontal",right:"100",feature:{saveAsImage:{show:!0},dataZoom:{show:!0,icon:null,title:null},restore:{show:!0}}},tooltip:{},xAxis:[{data:["v0.9.0","v0.9.3","v0.10.0","v0.10.6","v0.11.0","v0.11.6","v0.12.0","v0.12.5","v0.13.0","v0.13.2","v0.14.0","v0.14.1","v0.15.0","v0.15.2","2024-09-04","2024-09-05","2024-09-06","2024-09-09","2024-09-10","v0.16.0","2024-09-11","2024-09-12","2024-09-13","2024-09-16","master"]}],yAxis:[{}]};goecharts_4211100011286834090.setOption(option_4211100011286834090)<\/script> <div class="container"> <div class="item" id="5526230361037911490" style="width:900px;height:500px"></div> </div><script>"use strict";let goecharts_5526230361037911490=echarts.init(document.getElementById("5526230361037911490"),"white",{renderer:"canvas"}),option_5526230361037911490={color:["#5470c6","#91cc75","#fac858","#ee6666","#73c0de","#3ba272","#fc8452","#9a60b4","#ea7ccc"],dataZoom:[{type:"slider"}],legend:{},series:[{name:"Time (s)",type:"boxplot",data:[{value:[17.2688527,17.2786879095,17.306241167499998,17.4041588485,17.484358481]},{value:[17.30124315,17.3443548375,17.392648364000003,17.39937601,17.400921817]},{value:[17.211857327,17.239733539,17.369552456999998,19.309349198,21.147203233]},{value:[16.869445598,16.8831091805,16.9052222215,16.9424754115,16.971279143]},{value:[16.939166672,16.944832030999997,17.0269941,17.132707363,17.161923916]},{value:[16.787568998,16.8315816915,16.893251561,16.987963365,17.065017993]},{value:[16.72396799,16.752369618499998,16.828101089500002,16.8873099535,16.899188975]},{value:[16.660369445,16.672763546,16.7283587485,16.7756000925,16.779640335]},{value:[16.129276591,16.142885932,16.1965165605,16.251917477,16.267297106]},{value:[16.060027073,16.082007285,16.182163889999998,16.336555797000003,16.412771311]},{value:[15.995062631,16.022579776,16.0660190515,16.090445307000003,16.098949432]},{value:[16.01746703,16.0380534685,16.1072193805,16.2014604565,16.247122059]},{value:[15.976252237,16.034256796,16.097554759,16.1159046565,16.12896115]},{value:[15.987470908,16.0016205155,16.016933738,16.080167786,16.142238219]},{value:[16.532607825,16.5452847105,16.563189262999998,16.881020299,17.193623668]},{value:[16.479881014,16.502665954,16.533666263,16.62900985,16.716138068]},{value:[16.423171252,16.441621145,16.504929061,16.6098525195,16.669917955]},{value:[16.341978951,16.391692167000002,16.451512678,16.586766039,16.711912105]},{value:[16.289019898,16.330314504,16.395091880000003,16.876828514,17.335082378]},{value:[16.036750605,16.133467226500002,16.2384880865,16.2683861235,16.289979922]},{value:[16.378696994,16.3973146875,16.4173773595,16.434818190999998,16.450814044]},{value:[16.142905382,16.2029234945,16.344172238,16.43845111,16.451499351]},{value:[16.041188687000002,16.074042845,16.116735402499998,16.199873376,16.27317295]},{value:[16.030470013,16.0307176745,16.092408313500002,16.200292343999998,16.246733397]},{value:[16.065910667,16.073088113,16.102512074,16.308122605999998,16.491486623]}]}],title:{text:"Build from git context",subtext:"BenchmarkBuild/BenchmarkBuildRemote"},toolbox:{show:!0,orient:"horizontal",right:"100",feature:{saveAsImage:{show:!0},dataZoom:{show:!0,icon:null,title:null},restore:{show:!0}}},tooltip:{},xAxis:[{data:["v0.9.0","v0.9.3","v0.10.0","v0.10.6","v0.11.0","v0.11.6","v0.12.0","v0.12.5","v0.13.0","v0.13.2","v0.14.0","v0.14.1","v0.15.0","v0.15.2","2024-09-04","2024-09-05","2024-09-06","2024-09-09","2024-09-10","v0.16.0","2024-09-11","2024-09-12","2024-09-13","2024-09-16","master"]}],yAxis:[{}]};goecharts_5526230361037911490.setOption(option_5526230361037911490)<\/script> <div class="container"> <div class="item" id="13054923179277620692" style="width:900px;height:500px"></div> </div><script>"use strict";let goecharts_13054923179277620692=echarts.init(document.getElementById("13054923179277620692"),"white",{renderer:"canvas"}),option_13054923179277620692={color:["#5470c6","#91cc75","#fac858","#ee6666","#73c0de","#3ba272","#fc8452","#9a60b4","#ea7ccc"],dataZoom:[{type:"slider"}],legend:{},series:[{name:"Time (s)",type:"boxplot",data:[{value:[15.492267767,15.537656962,15.598390261999999,15.6264121815,15.639089996]},{value:[15.650760166,15.671703425499999,15.7151048005,15.770812251999999,15.804061588]},{value:[14.852958134,14.86501977,14.8810013405,14.8931282725,14.90133527]},{value:[14.31400574,14.355753318000001,14.397692003,14.425736765,14.45359042]},{value:[14.35144596,14.3572471125,14.415387963,14.4801612125,14.492594764]},{value:[14.381759094,14.3841688695,14.393273755500001,14.428915497,14.457862128]},{value:[14.339565197,14.3531269385,14.3765440775,14.420405109499999,14.454410744]},{value:[14.378538231,14.3873353385,14.396913721,14.3978054825,14.397915969]},{value:[14.238845149,14.244911769,14.253404819,14.277524626,14.299218003]},{value:[14.22679809,14.228435477000001,14.263374578,14.3827724535,14.468868615]},{value:[14.13903358,14.1408911495,14.145108754999999,14.168101545999999,14.188734301]},{value:[14.133816516,14.146462892999999,14.164614847,14.204102962,14.2380855]},{value:[14.361769336,14.3769600165,14.415723115999999,14.443028926499998,14.446762318]},{value:[12.260793219,12.2650480865,12.2933358525,12.3337137555,12.35005876]},{value:[12.281939207,12.287553922,12.3004641305,12.3455634185,12.383367213]},{value:[12.275007808,12.2936633385,12.313376448,12.341328418,12.368222809]},{value:[12.308940676,12.311060584500002,12.31322965,12.319975575,12.326672343]},{value:[12.210564469,12.2349481655,12.259575814,12.263607679,12.267395592]},{value:[12.210739064,12.2339512275,12.258683123,12.282509807,12.304816759]},{value:[12.244408484,12.255194617499999,12.2799159865,12.334141437500001,12.374431653]},{value:[12.208829704,12.236608276,12.267363342,12.316727336,12.363114836]},{value:[12.253529388,12.2605998095,12.270602664,12.292315927,12.311096757]},{value:[12.21049099,12.220898418,12.2511885065,12.2935617005,12.316052234]},{value:[12.205659516,12.206089827,12.2196354885,12.2661683395,12.29958584]},{value:[12.210231991,12.228380793500001,12.251400102,12.2612707945,12.266270981]}]}],title:{text:"Build with secret",subtext:"BenchmarkBuild/BenchmarkBuildSecret"},toolbox:{show:!0,orient:"horizontal",right:"100",feature:{saveAsImage:{show:!0},dataZoom:{show:!0,icon:null,title:null},restore:{show:!0}}},tooltip:{},xAxis:[{data:["v0.9.0","v0.9.3","v0.10.0","v0.10.6","v0.11.0","v0.11.6","v0.12.0","v0.12.5","v0.13.0","v0.13.2","v0.14.0","v0.14.1","v0.15.0","v0.15.2","2024-09-04","2024-09-05","2024-09-06","2024-09-09","2024-09-10","v0.16.0","2024-09-11","2024-09-12","2024-09-13","2024-09-16","master"]}],yAxis:[{}]};goecharts_13054923179277620692.setOption(option_13054923179277620692)<\/script> <div class="container"> <div class="item" id="8353803197768902519" style="width:900px;height:500px"></div> </div><script>"use strict";let goecharts_8353803197768902519=echarts.init(document.getElementById("8353803197768902519"),"white",{renderer:"canvas"}),option_8353803197768902519={color:["#5470c6","#91cc75","#fac858","#ee6666","#73c0de","#3ba272","#fc8452","#9a60b4","#ea7ccc"],dataZoom:[{type:"slider"}],legend:{},series:[{name:"Time (s)",type:"boxplot",data:[{value:[.650899492,.6523964655,.658800689,.6638303590000001,.663952779]},{value:[.657213787,.6601818975,.6644164535,.681001325,.696319751]},{value:[.736528462,.736735261,.740437798,.7439842235,.744034911]},{value:[.314718247,.315038518,.3159874875,.3207879305,.324959675]},{value:[.337808472,.339296461,.3427206585,.3463807075,.348104548]},{value:[.346641588,.348545288,.351479767,.356412562,.360314578]},{value:[.349030766,.349728396,.353861841,.360208421,.363119186]},{value:[.341807096,.34746901399999996,.3542399435,.3593542685,.363359582]},{value:[.33923451,.34177453150000003,.3461956205,.3510567,.354036712]},{value:[.333733921,.3383331035,.3447338605,.348417556,.350299677]},{value:[.259755287,.26067012700000003,.2619793065,.27068926849999997,.279004891]},{value:[.243655644,.2451727495,.2492055325,.251758495,.25179578]},{value:[.270725091,.271195062,.27197847750000004,.2744118185,.276531715]},{value:[.251419817,.2558147515,.26066653900000003,.26160987,.262096348]},{value:[.251763966,.2596776465,.26933357150000004,.2735091355,.275942455]},{value:[.273568326,.2759347785,.279667103,.28377576250000003,.28651855]},{value:[.261882637,.26678827949999995,.273682587,.27995758049999997,.284243909]},{value:[.261406236,.262768947,.266846427,.27031964750000004,.271078099]},{value:[.268700097,.2691963235,.273749157,.27962323050000004,.281440697]},{value:[.267704788,.272080878,.27670599250000005,.277147421,.277339825]},{value:[.268183042,.26915352849999996,.2715460705,.28354767449999996,.294127223]},{value:[.267298898,.2710662425,.27852612649999997,.2824355935,.282652521]},{value:[.258136359,.2630245765,.27221970500000003,.28036651300000004,.28420641]},{value:[.25022063,.2559405305,.261943287,.2641099785,.265993814]},{value:[.26233267,.2661329095,.270662382,.2756653765,.279939138]}]}],title:{text:"Simple build",subtext:"BenchmarkBuild/BenchmarkBuildSimple"},toolbox:{show:!0,orient:"horizontal",right:"100",feature:{saveAsImage:{show:!0},dataZoom:{show:!0,icon:null,title:null},restore:{show:!0}}},tooltip:{},xAxis:[{data:["v0.9.0","v0.9.3","v0.10.0","v0.10.6","v0.11.0","v0.11.6","v0.12.0","v0.12.5","v0.13.0","v0.13.2","v0.14.0","v0.14.1","v0.15.0","v0.15.2","2024-09-04","2024-09-05","2024-09-06","2024-09-09","2024-09-10","v0.16.0","2024-09-11","2024-09-12","2024-09-13","2024-09-16","master"]}],yAxis:[{}]};goecharts_8353803197768902519.setOption(option_8353803197768902519)<\/script> <div class="container"> <div class="item" id="6781028463788267987" style="width:900px;height:500px"></div> </div><script>"use strict";let goecharts_6781028463788267987=echarts.init(document.getElementById("6781028463788267987"),"white",{renderer:"canvas"}),option_6781028463788267987={color:["#5470c6","#91cc75","#fac858","#ee6666","#73c0de","#3ba272","#fc8452","#9a60b4","#ea7ccc"],dataZoom:[{type:"slider"}],legend:{},series:[{name:"Size (bytes)",type:"bar",data:[{value:38736177},{value:38796866},{value:39647930},{value:39558902},{value:51978978},{value:52050410},{value:53379146},{value:53379855},{value:57515638},{value:57533623},{value:57985213},{value:57988572},{value:59663644},{value:59650658},{value:59750164},{value:59746145},{value:59746145},{value:59727892},{value:59727892},{value:59727892},{value:59727892},{value:59764306},{value:59764306},{value:59771114},{value:59772059}],markLine:{data:[{name:"Avg",type:"average",lineStyle:{color:"#91cc75",width:2}}]}}],title:{text:"Daemon binary size",subtext:"BenchmarkDaemon/BenchmarkDaemonSize"},toolbox:{show:!0,orient:"horizontal",right:"100",feature:{saveAsImage:{show:!0},dataZoom:{show:!0,icon:null,title:null},restore:{show:!0}}},tooltip:{},xAxis:[{data:["v0.9.0","v0.9.3","v0.10.0","v0.10.6","v0.11.0","v0.11.6","v0.12.0","v0.12.5","v0.13.0","v0.13.2","v0.14.0","v0.14.1","v0.15.0","v0.15.2","2024-09-04","2024-09-05","2024-09-06","2024-09-09","2024-09-10","v0.16.0","2024-09-11","2024-09-12","2024-09-13","2024-09-16","master"]}],yAxis:[{}]};goecharts_6781028463788267987.setOption(option_6781028463788267987)<\/script> <div class="container"> <div class="item" id="4383527511627602225" style="width:900px;height:500px"></div> </div><script>"use strict";let goecharts_4383527511627602225=echarts.init(document.getElementById("4383527511627602225"),"white",{renderer:"canvas"}),option_4383527511627602225={color:["#5470c6","#91cc75","#fac858","#ee6666","#73c0de","#3ba272","#fc8452","#9a60b4","#ea7ccc"],dataZoom:[{type:"slider"}],legend:{},series:[{name:"Time (s)",type:"boxplot",data:[{value:[.032377259,.032420461,.0326357055,.033167115,.033743729]},{value:[.032329531,.0325520435,.033221134,.033337231499999995,.033459897]},{value:[.037711904,.0379956835,.0390438845,.039539858,.041359415]},{value:[.039231453,.0396169705,.039986077499999995,.0408663435,.041193433]},{value:[.03497111,.0352714385,.0362348325,.0370570535,.03755668]},{value:[.035054073,.0358894525,.036578508999999995,.036956223499999996,.037957535]},{value:[.034792946,.0350472265,.0354349215,.0359598125,.036279214]},{value:[.034237762,.034726629,.035415495,.0357254855,.036576806]},{value:[.034614933,.034895563500000004,.0353432105,.036029473000000006,.036706941]},{value:[.035041329,.0352666075,.035665243,.035997714,.036194709]},{value:[.034450247,.035882542,.037131557499999995,.038070637000000004,.038421028]},{value:[.034438116,.0349626455,.03578516700000001,.036839529999999995,.037715042]},{value:[.034541505,.034669517999999996,.034757433500000004,.0349524445,.035053838]},{value:[.03538756,.035810028,.0362677005,.036343263,.037046865]},{value:[.034442865,.0345617925,.035098505,.0357360355,.036295248]},{value:[.034417383,.034815038,.035387327999999996,.0359524485,.036513325]},{value:[.035027607,.035148523,.035414953,.035782760999999996,.036363694]},{value:[.035223893,.0356136665,.036227177,.0367909345,.037311477]},{value:[.033787112,.034646045,.035176758,.035822511,.036262563]},{value:[.034830671,.03512217,.0353661835,.035515098,.036167682]},{value:[.034372558,.0348119265,.035133324,.035779596,.036264919]},{value:[.034677807,.034970402,.035692485499999996,.036386361000000006,.037568874]},{value:[.033978951,.0350151325,.0353853715,.036065722499999994,.037210346]},{value:[.034982821,.035027208500000004,.0352306555,.036169723,.037012159]},{value:[.034893336,.0357948005,.0360086785,.036924956499999995,.03746291]}]}],title:{text:"Run buildkitd --version",subtext:"BenchmarkDaemon/BenchmarkDaemonVersion"},toolbox:{show:!0,orient:"horizontal",right:"100",feature:{saveAsImage:{show:!0},dataZoom:{show:!0,icon:null,title:null},restore:{show:!0}}},tooltip:{},xAxis:[{data:["v0.9.0","v0.9.3","v0.10.0","v0.10.6","v0.11.0","v0.11.6","v0.12.0","v0.12.5","v0.13.0","v0.13.2","v0.14.0","v0.14.1","v0.15.0","v0.15.2","2024-09-04","2024-09-05","2024-09-06","2024-09-09","2024-09-10","v0.16.0","2024-09-11","2024-09-12","2024-09-13","2024-09-16","master"]}],yAxis:[{}]};goecharts_4383527511627602225.setOption(option_4383527511627602225)<\/script> <div class="container"> <div class="item" id="1290314734383721111" style="width:900px;height:500px"></div> </div><script>"use strict";let goecharts_1290314734383721111=echarts.init(document.getElementById("1290314734383721111"),"white",{renderer:"canvas"}),option_1290314734383721111={color:["#5470c6","#91cc75","#fac858","#ee6666","#73c0de","#3ba272","#fc8452","#9a60b4","#ea7ccc"],dataZoom:[{type:"slider"}],legend:{},series:[{name:"Size (bytes)",type:"bar",data:[{value:118317850},{value:118367221},{value:115376730},{value:116151308},{value:145220202},{value:145303034},{value:147073992},{value:147116891},{value:170016409},{value:170050626},{value:169623823},{value:169627182},{value:173322062},{value:173308790},{value:173451893},{value:173540738},{value:173540738},{value:173486091},{value:173486091},{value:173486091},{value:173486091},{value:173522505},{value:173522505},{value:173529313},{value:173534612}],markLine:{data:[{name:"Avg",type:"average",lineStyle:{color:"#91cc75",width:2}}]}}],title:{text:"Package size",subtext:"BenchmarkPackage/BenchmarkPackageSize"},toolbox:{show:!0,orient:"horizontal",right:"100",feature:{saveAsImage:{show:!0},dataZoom:{show:!0,icon:null,title:null},restore:{show:!0}}},tooltip:{},xAxis:[{data:["v0.9.0","v0.9.3","v0.10.0","v0.10.6","v0.11.0","v0.11.6","v0.12.0","v0.12.5","v0.13.0","v0.13.2","v0.14.0","v0.14.1","v0.15.0","v0.15.2","2024-09-04","2024-09-05","2024-09-06","2024-09-09","2024-09-10","v0.16.0","2024-09-11","2024-09-12","2024-09-13","2024-09-16","master"]}],yAxis:[{}]};goecharts_1290314734383721111.setOption(option_1290314734383721111)<\/script> </div> </body> </html> ';const v=l}}]);
//# sourceMappingURL=7362.af931922.js.map