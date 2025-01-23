//go:build api_limits
// +build api_limits

package java

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestMavenSearch_GetMavenPackageBySha tests the GetMavenPackageBySha method of the MavenSearch struct.
// This is an integration test and requires network access to search.maven.org.
// It is not intended to be run as part of the normal test suite.
// Use this to validate rate limiting in [maven_search.go] and the ability to fetch package data from maven.org.
func TestMavenSearch_GetMavenPackageBySha(t *testing.T) {
	ctx := context.Background()

	ms := newMavenSearch(http.DefaultClient, "https://search.maven.org/solrsearch/select")

	// Known SHA1s to test with, using a large number of known good SHA1s to validate rate limiting
	// This is not typical but for Images with a large number of Java packages, this is a good test
	// to ensure that the rate limiting is working as expected and we don't silently fail and loose scan results
	shas := []string{
		"bb7b7ec0379982b97c62cd17465cb6d9155f68e8",
		"b45b49c1ec5c5fc48580412d0ca635e1833110ea",
		"245ceca7bdf3190fbb977045c852d5f3c8efece1",
		"485de3a253e23f645037828c07f1d7f1af40763a",
		"97662c999c6b2fbf2ee50e814a34639c1c1d22de",
		"21608dd8b3853da69c4862fbaf9b35b326dc0ddc",
		"a9cd24fe92272ad1f084d98cd7edeffcd9de720f",
		"eab9a4baae8de96a24c04219236363d0ca73e8a9",
		"3647d00620a91360990c9680f29fbcc22d69c2ee",
		"b957089deb654647da320ad7507b0a4b5ce23813",
		"bd0cd7ad1e3791a8a0929df0dcdbffc02fd0bab4",
		"0d1efd839d539481952a9757834054239774f057",
		"f6148c941e4ec2f314b285e6e4e995f61374aa2f",
		"502008366a98296ce95c62397b1cb7e06521a195",
		"92b2a5b7fb0c6a8dcd839d98af2e186f1e98b8ca",
		"64e6d9608f30eefbe807e65c148018065f971ca6",
		"095454c18fb12f8fcdbeae4747adfa29bfe6bf17",
		"0322a158f88b2a18b429133d91459dfa38bf9f55",
		"f18ebbe9a3145b9ce99733f5a0b7d505be9ae71e",
		"526df0db4c22be3eb490dab2b4ef979032e3588d",
		"521694be357010738e7bc612089df8fcc970a0d5",
		"50d87efaed036c7df71f766ca13aa8783a774ce9",
		"e8b2cbfe10d9cdcdc29961943b1c6c40f42e2f32",
		"3c0daebd5f0e1ce72cc50c818321ac957aeb5d70",
		"919f0dfe192fb4e063e7dacadee7f8bb9a2672a9",
		"8ceead41f4e71821919dbdb7a9847608f1a938cb",
		"a1678ba907bf92691d879fef34e1a187038f9259",
		"83cd2cd674a217ade95a4bb83a8a14f351f48bd0",
		"6b0acabea7bb3da058200a77178057e47e25cb69",
		"31c746001016c6226bd7356c9f87a6a084ce3715",
		"cd9cd41361c155f3af0f653009dcecb08d8b4afd",
		"2609e36f18f7e8d593cc1cddfb2ac776dc96b8e0",
		"0235ba8b489512805ac13a8f9ea77a1ca5ebe3e8",
		"ca773f9985c9f4104d76028629026c69c641923c",
		"a231e0d844d2721b0fa1b238006d15c6ded6842a",
		"8e6300ef51c1d801a7ed62d07cd221aca3a90640",
		"379e0250f7a4a42c66c5e94e14d4c4491b3c2ed3",
		"4b071f211b37c38e0e9f5998550197c8593f6ad8",
		"1f2a432d1212f5c352ae607d7b61dcae20c20af5",
		"a3662cf1c1d592893ffe08727f78db35392fa302",
		"78d2ecd61318b5a58cd04fb237636c0e86b77d97",
		"5b0b0f8cdb6c90582302ffcf5c20447206122f48",
		"0d8b504da88975fdc149ed60d551d637d0992aa1",
		"507505543772f54342d6ee855fa8f459d4bc6a11",
		"71abe1781fa182d92e97bf60450026cc72984ac2",
		"e0efa60318229590103e31c69ebdaae56d903644",
		"8ad1147dcd02196e3924013679c6bf4c25d8c351",
		"9679de8286eb0a151db6538ba297a8951c4a1224",
		"73b9a0e7032a5ae89f294091bc6cbb9a67a21101",
		"152f846d9f30a3e026530c2087ecd65c39bb304b",
		"73de3b1233c1da8fd46f9a4bd8ebec97890af9dc",
		"25d54640c4a17aa342490c4c63c172759361bf56",
		"eca76e00f897461f95bbb085f67936417ae03825",
		"802b5b3de0a38e71f07aa3048f532cd1246bc5af",
		"10d40ab670bf1fa53c925462f84f43507cf3b9bc",
		"2a14a2ff74f6ec3546b257889949630d3b2a0dbb",
		"357efe3f93c58bc4a10d40b1301045405b8a9f73",
		"570430f532b1e98c5d72a759ccbe7851099cee5f",
		"3174a146b81819fe2cd42e23081cd902ac743a8d",
		"940873068ea1383f4d962613cc1eca7c8cecc00e",
		"2116ab332c0bedfd038ad9d39c2e17219abf34aa",
		"527f9c5ccc6b76ad6e88ca571272a6a2ea535921",
		"04d21d5e6b71b2634dc67b36bf9b2defce7a7cc3",
		"37a5a4660941852c298e4caf4592b46b98ce512c",
		"780be6395b7c65d8d90ca2e1c3c2a46c46c5a154",
		"6251d68d3039f7b215b205f0e61cb2d732e5bc9b",
		"1d7efb089db2fe7a60526b8ff50b0c681fe1b079",
		"1f21cea72f54a6af3b0bb6831eb3874bd4afd213",
		"cd58e9e1b3ece090edd60a072f66b6cf52bce06d",
		"fcfd07e6ad0b5eadb0af1bddcc7b04097dacad7c",
		"e6fdf0f32f49d2a2380f5b458469052c272f8d9b",
		"324669468c32535f19bc4791fcaa34f2ed82200a",
		"ba584703bd47e9e789343ee3332f0f5a64f7f187",
		"17b3541f736df97465f87d9f5b5dfa4991b37bb3",
		"39e9e45359e20998eb79c1828751f94a818d25f8",
		"5353ca39fe2f148dab9ca1d637a43d0750456254",
		"603d37b2a108e2b437bb9b3b2ffb5962b4aa198c",
		"6000774d7f8412ced005a704188ced78beeed2bb",
		"537a3281dfefbd7939d27785732a2aafddd3abcb",
		"92446d8dfc8e57289e6120a7efc6932650ed3410",
		"eacefc2460e0ac5fe2ad48a9b0ffced5aea451b9",
		"4314021484adf9b32b3ae5421fac6fe0ed56e53e",
		"5786699a0cb71f9dc32e6cca1d665eef07a0882f",
		"2bd4f1921c78c2adffbe2eb01117c7936d0a0789",
		"de2b60b62da487644fc11f734e73c8b0b431238f",
		"e752540aeccb620f23c1e2f15c4c707254f6f596",
		"638ec33f363a94d41a4f03c3e7d3dcfba64e402d",
		"3fe0bed568c62df5e89f4f174c101eab25345b6c",
		"17773f342aabf0b177c9e3b8d8396d851cbfe64e",
		"1ae01f9be1cabf50ee735383a9fc3342e778c17e",
		"bf76d02e2be0dd8f99f106658ea7cacfa8df69d1",
		"f82b463a5c9eadb2a6667a1cb51b46d8d8d8d69b",
		"073e532b7cf87928bcd2512a0faf1151f8bd199a",
		"0912e12e4c7dc1c87ea8574065725a63342cf19d",
		"d52b9abcd97f38c81342bb7e7ae1eee9b73cba51",
		"dc98be5d5390230684a092589d70ea76a147925c",
		"47bd4d333fba53406f6c6c51884ddbca435c8862",
		"8ad72fe39fa8c91eaaf12aadb21e0c3661fe26d5",
		"54ebea0a5b653d3c680131e73fe807bb8f78c4ed",
		"19d5bfd402f91de0e670ef5783bf5c0a3f5ab478",
		"659feffdd12280201c8aacb8f7be94f9a883c824",
		"2b681b3bcddeaa5bf5c2a2939cd77e2f9ad6efda",
		"30be73c965cc990b153a100aaaaafcf239f82d39",
		"dc887691eab129c5728e26b095751fcadd36719d",
		"ddcc8433eb019fb48fe25207c0278143f3e1d7e2",
		"0ce1edb914c94ebc388f086c6827e8bdeec71ac2",
		"c6842c86792ff03b9f1d1fe2aab8dc23aa6c6f0e",
		"5043bfebc3db072ed80fbd362e7caf00e885d8ae",
		"f6f66e966c70a83ffbdb6f17a0919eaf7c8aca7f",
		"e4ba98f1d4b3c80ec46392f25e094a6a2e58fcbf",
		"4572d589699f09d866a226a14b7f4323c6d8f040",
		"bd1a6e384f3cf0f9b9a60e1e6c1c1ecbbee7e0b7",
		"3363381aef8cef2dbc1023b3e3a9433b08b64e01",
		"3833ca68f9f42fd11d4e0a036e9a3faae5d5f1a8",
		"4316d710b6619ffe210c98deb2b0893587dad454",
		"c22383d089321fd0c58a15c1c6ef5d24b5b5ee0c",
		"d858f142ea189c62771c505a6548d8606ac098fe",
		"66d618739859bc75ab9643b96a9839ac7802ec90",
		"e3aa0be212d7a42839a8f3f506f5b990bcce0222",
		"d25497d443d0843dbf2973e802c06722f2cb4578",
		"db2d83bdc0bac7b4f25fc113d8ce3eedc0a4e89c",
		"b706a216e49352103bd2527e83b1ec2410924494",
		"4aa0cfb129c36cd91528fc1b8775705280e60285",
		"4e1cce64b1ec11080a01172a0c296431d9469294",
		"e3fdd7fa9255bba0a206aea059cf133565c48cbd",
		"f0f717ed3495ed2e58d96e0084f73db0c7b3ba3d",
		"a79cf96a15f4b5376fae0024c0b0cd44cfa8a295",
		"49c3df840c2268479fb8f5cfd7df023bd6927bc9",
		"9d9d56fcae37f1b3d48d80f8b7eefabd3477569d",
		"09bfca4ee4f691f3737b3f4f006d0c4770f178eb",
		"0a1ed0a251d22bf528cebfafb94c55e6f3f339cf",
		"bc5b0c72a3755de7f3dca9f059aa19cc9d27a843",
		"a096cfeb58b927dde6b80ad295e564513514f9be",
		"b1e952300954b6d33911ba29a984455fcc3f1024",
		"5774f912db3dca1e9049af15cce6a4f7845a173d",
		"a2f8cf63192ebba929451a221cc382bc0ca5abb7",
		"13e3663d5878001666981eb5ef6efb22fa6799bb",
		"9697b9e1667b4f2daa9ea454b4a0e0f905585c8b",
		"32088dfde15a3f8ad4f2547cb083777afddc12d5",
		"fabcda911ebc80e3a9b6064863da4f2e5094814f",
		"2c3591cf5e2f5de644aae09a73a896f0c7964f43",
		"bce88f90c3341ed14df2ce3919f253334cd834f2",
		"df7bbc5a4c8304aa8aed34cb67e339035ac2c34b",
		"c305f6229dde8f3946de5574ac9779309073f2e3",
		"ad63993db3525be5e290e0ccb3d5122c01bd356d",
		"24b20d4f91c894e19947389d3040adfc174a6af1",
		"58c3d2641b48a9db2e29009f42077dcd70f7e351",
		"8450fb3261e7ec1d734c2b11ca4d875fe82386eb",
		"9a296a2da46d296f3d0b78d3941ec468c64ba3e6",
		"bd7b0f03050125e8dd8bd9498e34561e1e88db03",
		"b6104ad646d672770561918073f1aaacb7c7b341",
		"8c177eb55da21bee1cd654d66241b98fb0e44c86",
		"1c45fbaa5f4d66070b7f1ee5e4653aadb14aa97d",
		"0ed231cd84006f5fdfda7671beae2b9b41a2dafa",
		"8ed4fee000f82e6248f7f8cfdd11d53fe03f98ad",
		"302ebf7b124c9a037333a9b81a5f2ce0880f8a29",
		"eba91bffe866a695d145c5e1692509f92de5b23b",
		"5a1f4a878b75dbcfbf0d4ae783bf1c1229309470",
		"5351a31139b9b5e3f8d50252ac081249b1ad00fb",
		"fdc6f7632078dc5b570f9120d9ab07892e784554",
		"4a0126da8cf7794e913a13e3f8f4ab62ca5e2981",
		"4a3df17312a2ab95a4d75396065079aebfb2a1e7",
		"51fac22c802ae94247664efd95a1e60d138a278d",
		"c6dd14eb5a4abfcf1c8dbc7187c2ec3b8d9be1f9",
		"96d70f8f82a534438b938f86b3a6682eb34824ca",
		"e714165da098686f600d75b914448fdd4a057d60",
		"60ba0670d68758e893870079916954a7f01afe23",
		"320e7d1fdbab2bffb8138d66c24724cc24ea654c",
		"404840df034905ae2b5a9c922639e1d9f694516d",
		"63f0c49628c9695704d0014409f030d82bc10f70",
		"68186dc73e3d123999ebb93a6d5a5d0bbb4d4e91",
		"9ba2ed9f74f5122f25113cd6d5e14fbc442c867f",
		"993a5608c4942b5b81a6c14fd78779d024e6ed41",
		"f07ec0309a1e37629f097408e1cb75f7d0ea58c5",
		"1ffcac9e1bbd3d00db1e2089d8e915f20c0ac568",
		"9450776e99a5a1b413b98cb095f6fe7f81935c3d",
		"c72b35c5dea306de35ad0ff207eff4d14b37b880",
		"5e4e7abcdb8f4101b9aa0ba84658be21c445b1d5",
		"ef74ce50e19736bf72341a572c1ad6fd2ba6c3fe",
		"55a266187baa9d1c68447ff6ab404a4324de7935",
		"8a65a223354726586d95f45aa8f6175ca23b784c",
		"2ddd12523600e8b80d2be0bc003cd447bd2751d4",
		"19160c71c598866e9c96af667045c886c8dc9b48",
		"163372f10bf5f028ccbb122eebc9cd2deb30b094",
		"566ab030e0a0f010dfe0d185b0804b53817db7ec",
		"0c74d5b6c2ef578266361a58ec7c848cd844f2bd",
		"452cd7f4850757ad76710cea53bd9ad8d181d5dc",
		"98cbe204421b538fd2fbf4a1ce689f8398bd2ced",
		"b917f21f99eeacf49f55e8fd089b93119c7dbd9b",
		"bd4d8f4a02886a26b60c76048547a453691fcec3",
		"1dcf1de382a0bf95a3d8b0849546c88bac1292c9",
		"799748e42a644db85394db066af658809f89c523",
		"c693557ee87e311340eb0f8a811b8bca027af421",
		"912b86862ad070dd3d21f51e05e361eba1f515da",
		"67b085271fd9cc0a61eb04fcaf288ad35b2e7995",
		"a41a8b5641dad26c7601ea93818611b4a6465058",
		"5ede807d3bcdace2e25d5614382bfdf1663012e5",
		"8275c3b8829eb16a54fb49ceda2f6fbb44546c26",
		"5a2b47396587b499575782b60cb223a830bc86d7",
		"e113ac14fb2b70c1510f92ea2a0405ba4da01f5c",
		"10e53fd4d987e37190432e896bdaa62e8ea2c628",
		"286c93b65ab3c3a0a257b0a6ebdd99c06c674c88",
		"7b93e7e3c64b837b69da7497fdf4c28b677625bf",
		"0bc23b2c7e6419d3cd7e108d6942b9431bf5c25c",
		"0a5f0e4a16f5b12cde3df1ea413852aeaf176176",
		"44984c2480ac8aaef4a660a06565aa76c577238c",
		"5878d0f20e7cc521a437217dd21c3a84788d3a53",
		"73120785e720701d1142d97bdc72bf5d6b5af4bd",
		"fee8f41ab7f59597e35d8a6eb01b9edc9b04d51e",
		"e0feb1bd93ad9fb1e064706cff96e32b41a57b9c",
		"3dc8cea436c52d0d248abe9648b0e4f1d02bd500",
		"3e224b1b9e18dd28c89a764b1feea498ba952579",
		"09d6cbdde6ea3469a67601a811b4e83de3e68a79",
		"3251b36ee9e9c3effe3293f8d7094aa3841cad55",
		"252e267acf720ef6333488740a696a1d5e204639",
		"789cafde696403b429026bf19071caf46d8c8934",
		"4a4f88c5e13143f882268c98239fb85c3b2c6cb2",
		"8046b9d6b423f24457cfb20210d0ee8abc98e22c",
		"bb4eda2c61102759e7c03ab12ff7c19547e20cbd",
		"52da76c0c8190be88281aec828efd44df176ab34",
		"878e2200222f5e11137d5bfde325a5db30687592",
		"1789190601b7a5361e4fa52b6bc95ec2cd71e854",
		"44f8a2b2c0dfb15b4f112e22de76d837b89bd4d6",
		"aaf681a518ce5c9a048328b86ba5b9c5123375aa",
		"dbd77d2e6c54ed9fafc83f1cf6f48342250996d7",
		"532fd1449686690273222ebd5cb86c233ed19f58",
		"60de19e6c8e44b1a78acb0dd73722b2feaa7ccfe",
		"85289261815e7d2fb1472981652fce50ae8cfc42",
		"71b610fca525744bc70eb96c9f9113cddbc38f4f",
		"2977cca2c82e3c5336805ebb6226c14137585b54",
		"1d2d1a5ed9cfc58d0a7bdc1d9dda5ecb9987da9a",
		"d339db49e637d2a8122a67ad846a294124f1a2f3",
		"02f16015ed9e2689e10f86f1b7c3522e541c0c75",
		"0e7eab15d6b4184921b82fbca6f89dcb60ea972e",
		"362e2295d95b2e2797457760eef1f172d07d7417",
		"c067143934cb76530adbb8fd4e2df1ab737a16e0",
		"b3add478d4382b78ea20b1671390a858002feb6c",
		"907df2bf39d70510951b7bafbf661f286eed90a5",
		"6e5d51a72d142f2d40a57dfb897188b36a95b489",
		"00f6db9a5e6fe2374b5a494b08388c2b6e0792d8",
		"eeb69005da379a10071aa4948c48d89250febb07",
		"af799dd7e23e6fe8c988da12314582072b07edcb",
		"3b27257997ac51b0f8d19676f1ea170427e86d51",
		"90ac2db772d9b85e2b05417b74f7464bcc061dcb",
		"451bc97f7519017cfa96c8f11d79e1e8027968b2",
		"066aaf67a580910de62f92f21f76e3df170483cf",
		"d5e162564701848b0921b80aedef9e64435333cc",
		"12ac6f103a0ff29fce17a078c7c64d25320b6165",
		"21f7a9a2da446f1e5b3e5af16ebf956d3ee43ee0",
		"81065531e63fccbe85fb04a3274709593fb00d3c",
		"09ca864bec94779e74b99e84ea02dba85a641233",
		"d635e3eed4beb74213489ff003ca39dbe47ea44e",
		"b75e5e9feb70f599a6f6232e71bd5b0030608179",
		"02419d851c01139edf9e19b81056382163d9bfab",
		"57b7ba0ca94313c342b03bd31830fe4a8f34bc1a",
		"a68959c06e5f8ff45faff469aa16f232c04af620",
		"70b332574395cde2c56db431b619be9823407aed",
		"45c3bb7696f29655189abb78ec1c97f511643159",
		"99a1348743f3550dd4524408725efab8eb319960",
		"acc766c65cd4e94a5e1fab6a2f85148dfc8613d8",
		"28bdcdcceb92a1ac450b8b6a3d3d0627d839054d",
		"9ea12cb2c426d521b7c4cad5b02ce18e5b614d4e",
		"8347ef8861b75bbffcaebb706a0ae296daabc20e",
		"e844c4278ecba985c08e0dea1181343a07c04c3e",
		"4b151bcdfe290542f27a442ed09be99f815f88e8",
		"1d7200e19d1ffdaf6927ff0be701724c85be07d7",
		"1861e32c3c484a1aa5f55ab109e08dd1b32c6fa2",
		"34c56f43fd3255fc239ffe33d0fbfb8195be6a24",
		"e5f6cae5ca7ecaac1ec2827a9e2d65ae2869cada",
		"0c900514d3446d9ce5d9dbd90c21192048125440",
		"56b53c8f4bcdaada801d311cf2ff8a24d6d96883",
		"de748cf874e4e193b42eceea9fe5574fabb9d4df",
		"4bafcb5aacb1abc193698a13ae99394e09a25101",
		"687cede1a44f70c7741abfab6ee2aa53dd2bfb54",
		"34d8332b975f9e9a8298efe4c883ec43d45b7059",
		"698bd8c759ccc7fd7398f3179ff45d0e5a7ccc16",
		"2872764df7b4857549e2880dd32a6f9009166289",
		"164343da11db817e81e24e0d9869527e069850c9",
		"e1c6222f2fa8d05d7825d8e9af7b9bef089c0b5e",
		"127fc12785c42eeff7da15abca690655add7c710",
		"53c041061964825372701d75b96a67e82bc3b6da",
		"ba4bfac0366399080e878cb5c41023c3eb7f7328",
		"61ad4ef7f9131fcf6d25c34b817f90d6da06c9e9",
		"4a3ee4146a90c619b20977d65951825f5675b560",
		"67613a3a83092588b85b163b9aeb3e87ec46b4ea",
		"c197c86ceec7318b1284bffb49b54226ca774003",
		"ba035118bc8bac37d7eff77700720999acd9986d",
		"c85270e307e7b822f1086b93689124b89768e273",
		"2042461b754cd65ab2dd74a9f19f442b54625f19",
		"04669a54b799c105572aa8de2a1ae0fe64a17745",
		"48d6674adb5a077f2c04b42795e2e7624997b8b9",
		"15177b3e1c91529ba02c056035d71463f9e66a03",
		"241aa26c61f638b7e76787558bb1be49984f2a0d",
		"4605d2f4267388d02d810a3cea448a48371435a3",
		"d1cce505a1dafd5b5d842ec8f91105ccca7d5e4d",
		"0a10ae77a57942352f3d2abe5d58b199b1f83d33",
		"dc6c49c40b1d5acf3ee58784ec34360806f48a22",
		"6d9393fd0ea1e1900451b5ee05351523725392bb",
		"8a603c50591cd2ba1039afa3f28540b0f43c82c5",
		"b4bd511b8cff2cfd83faf37f48b6e256dabcfc6c",
		"c02f6844c6d27c8357257eab162250b54d38390b",
		"46cb8029e744ac128d4ce58c944ab21ab7ef3e25",
		"e27eb84680badf95422bf5798faf8b0b7fa332f4",
		"a974a4972c0ecd2206b045e387bea4713a5451b6",
		"e3480072bc95c202476ffa1de99ff7ee9149f29c",
		"74548703f9851017ce2f556066659438019e7eb5",
		"562a587face36ec7eff2db7f2fc95425c6602bc1",
		"f3cd84cc45f583a0fdc42a8156d6c5b98d625c1a",
		"f48473482c0e3e714f87186d9305bcae30b7f5cb",
		"288f60226c596849c3c57926e8421c83b5abbe87",
		"5eacc6522521f7eacb081f95cee1e231648461e7",
		"5eea182d6651a7257bc8c3614507e1540c766fc2",
		"8d49996a4338670764d7ca4b85a1c4ccf7fe665d",
		"48e3b9cfc10752fba3521d6511f4165bea951801",
		"0422d3543c01df2f1d8bd1f3064adb54fb9e93f3",
		"878a02f3ab98d37206c9852c025a46b86dc882d0",
		"c8de82b962142a5f3d408ecc3920642b166de028",
		"bf744c1e2776ed1de3c55c8dac1057ec331ef744",
		"85262acf3ca9816f9537ca47d5adeabaead7cb16",
		"934c04d3cfef185a8008e7bf34331b79730a9d43",
		"60a59edc89f93d57541da31ee1c83428ab1cdcb3",
		"935151eb71beff17a2ffac15dd80184a99a0514f",
		"3cd63d075497751784b2fa84be59432f4905bf7c",
		"8531ad5ac454cc2deb9d4d32c40c4d7451939b5d",
		"3758e8c1664979749e647a9ca8c7ea1cd83c9b1e",
		"dd6dda9da676a54c5b36ca2806ff95ee017d8738",
		"40fd4d696c55793e996d1ff3c475833f836c2498",
		"ef31541dd28ae2cefdd17c7ebf352d93e9058c63",
		"d877e195a05aca4a2f1ad2ff14bfec1393af4b5e",
		"39e8f0d32258f304928b29bc7e1f7d85fa5ae218",
		"f9414e9ed1a16132c5d3467991a3bebc4367a1bc",
		"984a623e0c0dfec82cb1cd390ee1aab51fa02bee",
		"d60d3f8ccefd848d551d25bcb7f3e9251333648b",
		"e9cb36b5954d1af1593bbc4fecadfa5cb170bd44",
		"058e7a538e020b73871e232eeb064835fd98a492",
		"6f14738ec2e9dd0011e343717fa624a10f8aab64",
		"7aaae28e06aafe63ac94f7c6dee81135b815db92",
		"9b1f3cf3fdd02d313018f1a67c42106e6ce9f60d",
		"21c5319c82ca29705715b315553a16f11b16655e",
		"fc5cdccdeeedb067b8b2a3c7df2907dfb7e8a1b9",
		"44df9a1310c1278b62658509aca3ca53978e8822",
		"fbdcb39db6a6976944a621fe11bf1d2ff048d7c2",
		"1a01a2a1218fcf9faa2cc2a6ced025bdea687262",
		"056dcc8480ecd2c03ec004aa76278d1f2d621561",
		"b33d6d9045da8f0b317162facabcd1dc9ebf751d",
		"be8f9b519d692dfd1e2726ee9e26573dabc99e70",
		"e522a857d234e5ade679abfae807bcf4ecdd6f2e",
		"533e7cbc5efc1d58d14cefb68904cd3af47fc316",
		"fdeb0e5beb9ddbfb49b4aec3daa55d71e0cf1956",
		"966952ede72900ddaa20888beddc86a5a002cbd3",
		"b22ab0397e893d9c092ade34a8e826ef576b285c",
		"65ce500f7cad946ce0e172c6ce90319caa29e787",
		"c8645a939af24f227e4123b89af14264176f7c60",
		"62bbe12f1a737d9b96358a9964466ffea6a6487a",
		"696cf9f9160e3a0ff208db614af118915cbdcf2d",
		"2f2695de1ae62d84ca8336c7e6ddedc80aa2e521",
		"59d5e3e86e2583fba0cf04d4f126a5205a24c4b6",
		"d60bb33b97b968b555ce829961d41971ff826415",
		"1200e7ebeedbe0d10062093f32925a912020e747",
		"88e9a306715e9379f3122415ef4ae759a352640d",
		"0a1cb8dbe71b5a6a0288043c3ba3ca64545be165",
		"a240efb690601cb1ef02a6778a23e450559b0bef",
		"0e7c45d7667feb56e5664247a882451c3d438def",
		"eaa7ec646db93f0096ca8100c361018c2608319b",
		"cef76bca08c1f437150890d1b8bf430a66ebe42a",
		"21743fe8af7bb684d5ebbd4075f397fccc30d158",
		"006936bbd6c5b235665d87bd450f5e13b52d4b48",
		"698ce67b5e58becfb4ef2cf0393422775e59dff4",
		"a698fd936b13588c6747b182bcfdc13885a8ca43",
		"357a3836bb5da16f314f3a1e954518e5468cd915",
		"37fe2217f577b0b68b18e62c4d17a8858ecf9b69",
		"5e303a03d04e6788dddfa3655272580ae0fc13bb",
		"c9ad4a0850ab676c5c64461a05ca524cdfff59f1",
		"cc5888f14a5768f254b97bafe8b9fd29b31e872e",
		"63f943103f250ef1f3a4d5e94d145a0f961f5316",
		"516c03b21d50a644d538de0f0369c620989cd8f0",
		"25ea2e8b0c338a877313bd4672d3fe056ea78f0d",
		"59033da2a1afd56af1ac576750a8d0b1830d59e6",
		"2ca09f0b36ca7d71b762e14ea2ff09d5eac57558",
		"3ff3baa0074445384f9e0068df81fbd0a168395a",
		"4c7018119aeb66335746e6748456c821e304d3a2",
		"75a75c47eb912f3fd06df62a9e4b3b554d5b2bec",
		"2e617bd795b3b55b2ec23543721665a2b1c77b9c",
		"a0f0c4de6cc321130252e86658c21b2e1b6af008",
		"7868b29620b92aa1040fe20d21ba09f2506207aa",
		"a82d2503e718d17628fc9b4db411b001573f61b7",
		"e358016010b6355630e398db20d83925462fa4cd",
		"82357e97a5c1b505beb0f6c227d9f39b2d7fdde0",
		"66eab4bbf91fa01ed4f72ce771db28c59d35a843",
		"eb91bc9b9ff26bfcca077cf1a888fb09e8ce72be",
		"c56ffb4a6541864daf9868895b79c0c33427fd8c",
		"1e39adf7c3f5e87695789994b694d24c1dda5752",
		"93d37f677addd2450b199e8da8fcac243ceb8a88",
		"d54a9712c29c4e6d9d9ba483fad3d450be135fff",
		"a4c3885fa656a92508315aca9b4632197a454b18",
		"4c1fd1f78ba7c16cf6fcd663ddad7eed34b4d911",
		"389b730dc4e454f70d72ec19ddac2528047f157e",
		"7d1b5b69a5ea87fb2f62498710d9d788d17beb2b",
		"b8af3fe6f1ca88526914929add63cf5e7c5049af",
		"dafaf2c27f27c09220cee312df10917d9a5d97ce",
		"7473b8cd3c0ef9932345baf569bc398e8a717046",
		"67f57e154437cd9e6e9cf368394b95814836ff88",
		"3c2af9d14e43d46b541ac1a0cdd7be4980aeed84",
		"aec7142dafa1f96154018eced507854bb544cf41",
		"319d3da49ca42fca687de2accef1d22fba786405",
		"7577f792244cae44227e675aafcf6597a2eeb00d",
		"9df166a4a89314f5281b37e524bb366d7ffadf23",
		"a0b8af84c1ddf5d9dd7d1eef8a19df527864e8cd",
		"ba91c4fa57b566baff698a3354db2b8af8626d3c",
		"7d5c94b0fb6384b91d963d6d398468d96bb4983f",
		"4c54840ac217908029e77a96336c03901a6776d5",
		"6ac92abcc06bb8d52d8179889c6b1173ba7bd027",
		"8bc95edbea781cc09dc6755f570b72a993df1679",
		"a9992b9918cd8582e2fef748a61ec1c46894b13f",
		"bb8b60297070d7c352a06c6bbc3854cfac26d46a",
		"4185a5807b9fdbdcbee813f8e82ceb433ad75c68",
		"9bc1a58a9472452100f873059683a9a37e37063c",
		"4d1701b9c993f5edfd232ea06a6f8ba540113b59",
		"5d5c7c1b342c89b4b0d28cd99572827cbe3e6f15",
		"d641ffaf2a3a84c8c85d24850b916c5baf547a38",
		"c5bf5e88b285eda01f3d2044c76a6f5651dfa4c0",
		"6b7aca9462a226dbdef319e6875523e322c6f80b",
		"b9740e040f5fc06920f38d9fafd966bfd3cabe1e",
		"830a7d5e13edb5f6f81c36877edf68b55a4182fb",
		"751030d0b6c06337bb2870cd174ec83ed8417b3e",
		"b8957915e5d02d9e341eaa07a90019aeb90d546f",
		"82cfcd48e0c239ddbdf4fa122b8715275b761de2",
		"97c73ecd70bc7e8eefb26c5eea84f251a63f1031",
		"5d1abb695642e88558f4e7e0d32aa1925a1fd0b7",
		"0e5af3b6dc164eb2c699b70bf67a0babef507faf",
		"f08a912ce02debbaa803353686964b3c5fcfdb53",
		"8625e8f9b6f49b881fa5fd143172c2833df1ce47",
		"b421526c5f297295adef1c886e5246c39d4ac629",
		"9be9bb9b3a1638dcd948edd6179bd8ee4ffcc137",
		"bea6fede6328fabafd7e68363161a7ea6605abd1",
		"7183a25510a02ad00cc6a95d3b3d2a7d3c5a8dc4",
		"af40e34de7c30e4fef253024a88257f6dddc547a",
		"d2cac68225a25a5486ea848af95680573ee3d393",
		"d952189f6abb148ff72aab246aa8c28cf99b469f",
		"87fa769912b1f738f3c2dd87e3bca4d1d7f0e666",
		"79d7792942fa009316de2d7d1a4d7e8b33548947",
		"6604030f7da573a8c00641f9c7deef6c143b6022",
		"2746f9ec96f9ce3a345b11f03751136073f7869f",
		"f73773fc39d43df7661609b9f7a733ddfd091af7",
		"be8a20787124cf52c56c5928ef970df2d8a26f51",
		"8ec1dce97ba5b616e165068225bba873179482e9",
		"dff1e225fe6bfdf7853663bc48831e9714bf035e",
		"ebe2549568386d5c289ec0eb738172f1a0445259",
		"9d5fdc88f91586bf5d1afa13b9a77302c39b5e7c",
		"ab7c7c3c823cb2f8fb1b54fdc82b3e133e8e8344",
		"bc34429b8d1a620c58639f376bee9ba425a035d3",
		"0bc8d9f00bd34806bc82d01390855ef9dcbea85b",
		"a1e978879d35af3590549437b80679b5c00f27d6",
		"1000c919125bb13f265b101341c34bb5af814fd3",
		"488e5cfdd4d2d30b161fc45819a82a6984eb0f99",
		"4b986a99445e49ea5fbf5d149c4b63f6ed6c6780",
		"64485a221d9095fc7ab9b50cc34c6b4b58467e2e",
		"bf9e9aea47c3d112929fc56abd75a48d31914fda",
		"73334ff5470db03e5b793ce1d5854642b2c21799",
		"7fd8a65d950d0e77dd39cc4ce2776ff9673ae470",
		"d4c0da647de59c9ccc304a112fe1f1474d49e8eb",
		"ccf79a1a63ef35de038a4226a952175c4e9f4f59",
		"5fb53c92da84ebeff403414b667611d6bcd477cf",
		"ef5bccf2a7a22a326c8fe94e1d56f6f15419bedd",
		"311d38cf15ec7f5c713985862632db91b7a827af",
		"e2d5e96ea4bbd4fc463dbb76d07dd8aefac05e3c",
		"61625cf2338fe84464c5d586dbba51d4ff36a2b8",
		"9d8cd3ed749f2c2f846e0c58c485c8a0d5d5181e",
		"2f23beded3e46a3552fc3c1a0fdfb810c24d8f97",
		"54bc99d2b886a868d79d537ee5e7829bb062fbe4",
		"52fd60d5dc3f0fb3ed5c19b63f6f2312cd1f6add",
		"8a8ef1517d27a5b4de1512ef94679bdb59f210b6",
		"f6ea1e9c0a0acf137a8a4c5353bc97ead6b82cf7",
		"ea93fbd2137c797ed8a686737e8bdfeead20f1b1",
		"18ed04a0e502896552854926e908509db2987a00",
		"2a9d06026ed251705e6ab52fa6ebe5f4f15aab7a",
		"c2ef6018eecde345fcddb96e31f651df16dca4c2",
		"93cc78652ed836ef950604139bfb4afb45e0bc7b",
		"dd44733e94f3f6237c896f2bbe9927c1eba48543",
		"ed90430e545529a2df7c1db6c94568ea00867a61",
		"3ad0af28e408092f0d12994802a9f3fe18d45f8c",
		"9da10a9f72e3f87e181d91b525174007a6fc4f11",
		"d186a0be320e6a139c42d9b018596ef9d4a0b4ca",
		"62b6a5dfee2e22ab9015a469cb68e4727596fd4c",
		"84ede759015e7480ca8e6ec2af6e2f596aa92dec",
		"f3085568e45c2ca74118118f792d0d55968aeb13",
		"84d160a3b20f1de896df0cfafe6638199d49efb8",
		"6915c9c6966bf1482ac93236453013535e8c5d80",
		"bd1236bebd1ee50c8d9206e69b1986fac9532a49",
		"70cff2bc010d0c047cf5b167b2c600e42f6863ab",
		"6610ef4a025fcea2a5958724b9493a1b081b8f66",
		"ca018bb3db661230fbf51bc2b3b1559ac7987040",
		"313e89f2da215f0dcc54a638c5749c4ee959e74a",
		"c0dc8a542fd18d372a2ee67a203f2cfe0a345a05",
		"b31c6944d9cfd596b6c25fe17e36780bfa2d7473",
		"3a7aecd4bcaf75c7b0b02c26ea6ceacf3e8f5f4d",
		"1fd80f714c85ca685a80f32e0a4e8fd3b866e310",
		"baf7b939ef71b25713cacbe47bef8caf80ce99c6",
		"118f166726472bd5b5578817503ed0992c9102e2",
		"4c62b2337352073ff41fa9a9857a53999d41b49b",
		"3addc6860c44edcf28c262489f23276b84e11812",
		"0df31f1cd96df8b2882b1e0faf4409b0bd704541",
		"a224b43863a0679f153bec24e1d329c63f1ed234",
		"700f71ffefd60c16bd8ce711a956967ea9071cec",
		"fa9a2e447e2cef4dfda40a854dd7ec35624a7799",
		"6d62b9b4db6228122a5f1cda81b06f156afb04ba",
		"14f50cd1c2f5d29d9b070746c1fcae59b68ca26b",
		"d3e1ce1d2b3119adf270b2d00d947beb03fe3321",
		"2f4525d4a200e97e1b87449c2cd9bd2e25b7e8cd",
		"b0b14b3d12980912723fb8b66afb48dcda742fcb",
		"bc28b5a964c8f5721eb58ee3f3c47a9bcbf4f4d8",
		"49b64e09d81c0cc84b267edd0c2fd7df5a64c78c",
		"8bf9683c80762d7dd47db12b68e99abea2a7ae05",
		"5600569133b7bdefe1daf9ec7f4abeb6d13e1786",
		"66a60c7201c2b8b20ce495f0295b32bb0ccbbc57",
		"3c13fc5715231fadb16a9b74a44d9d59c460cfa8",
		"c05b6b32b69d5d9144087ea0ebc6fab183fb9151",
		"b7ce164e9e75be4b5eb42fd89c9c53ebecea6729",
		"abc7bac20e8b15d5aa38d1c9af5bed4e0ffc7748",
		"928c299530ecce5c25dcf62f72f6aa901d6baea4",
		"87b2ed1c62d42fd9fbbd154095f2387c6a18f880",
		"67336cfb9d93779c02e1fda4c87801d352720eda",
		"074b9950a587f53fbdb48c3f1f84f1ece8c10592",
		"132630f17e198a1748f23ce33597efdf4a807fb9",
		"00b5ec860e174d7a2edb2b46523cdc5401513cbd",
		"3b17774c8087e239542afe1c7976c16c5446af26",
		"cc71779727e9051e59c8a242b4157fc1d3172caf",
		"9aab69982e4a9b91a86743f73dc48db30daf9265",
		"ff0cd44f590a80c5c87aaa85a0d2bab2d350bc4a",
		"7b6c7f676d78f988b01e9841ab18d389886ffa26",
		"5fcf76dda71647a65d6fafbab2bf03065bf3d52d",
		"c63eaf104979cae41c9c5b2ef1a9bbe5d1c05480",
		"a7efb5dce8081d1d96445355d55f05e6e825d41f",
		"89223f29832931516d6c1f00a9ef2263b8674f5a",
		"5506a7066998a2be47b86e28d061863a475a7ca8",
		"a2e83c6e6ad2086f97277540d9d9ef4aebb74a28",
		"da50b1b4177cbadb977d52aa70011713f37a2156",
		"e7d90c28cbbe26b9a31fa8a136c209418ca3c9ba",
		"0ce981aa4a840f84b670bbb3dfd77cd3be87ca84",
		"1c973b3f5c13399e1194724abe421e230c572206",
		"9074f509fbc3df3ad104eca5427d03eece453246",
		"34994ed5371f31eeaa68b294d7f729934280a733",
		"57001fa0b6622767e63c1b9cd2e6db666d180caa",
		"6a999a46cb630f44f1a77ac39213fad57a8c1492",
		"2c3fe5d2e5941e50947ff59c50d201d3968fac02",
		"1149e08b436cca632ddfe8cee39918f23b50dc6f",
		"0b813b7539fae6550541da8caafd6add86d4e22f",
		"ef65452adaf20bf7d12ef55913aba24037b82738",
		"b260ca7a23bb0d209771db7aae35049899433fe3",
		"b4ac9780b37cb1b736eae9fbcef27609b7c911ef",
		"86ed42574cd68662b05d3b00432a34e9a34cb12c",
		"a483da1de9cb174ca327059e9fd8432b0e8666b3",
		"6eb2c27f1b7d048a6912a42a0637e470cdc46562",
		"1d3f5d1fd272883cbc26f3d7fcf9ba58f66d48e0",
		"5204ace0d7b8410a5fb73c17a20a69e616215131",
		"60164baf43273401883c7c0b53b0bc4359b9e94f",
		"7fbf34d79ca897acf21061c2e24b607b090be1c9",
		"5ae5c9ec39930ae9b5a61b32b93288818ec05ec1",
		"f90394c695d47b16f608be5366373eec768597f1",
		"e0d6c62cef4929db66dd6df55bee699b2274a9cc",
		"fff73bb736a3ebf11974ba2ded176f16a1976f0d",
		"a4ad886bfdbd1a872bdb3b25a9893994b78adf11",
		"010245305f4faef0ed473552a58f83d281754e77",
		"b82b13e45d9372296362e0d6dc481f6a0f2ce0c7",
		"e6396ecb39ea2c91dc9901213da1d29b8ae1798c",
		"1ed09f94667962983cce7ec6c7a1df5c0881e08a",
		"ae1536faa3401b0c62f93e29fef9ffcf134a616a",
		"a26d3c16f32cf21cbe24c0d7dc37132c407608bb",
		"d716952ab58aa4369ea15126505a36544d50a333",
		"2949632c1b4acce0d7784f28e3152e9cf3c2ec7a",
		"323964c36556eb0e6209f65c1cef72b53b461ab8",
		"3864a1320d97d7b045f729a326e1e077661f31b7",
		"6f29a4f68e4156358f64f6a060c5e55ad42f5231",
		"e1e99c956a36e619398f9e94d775f51a85c26770",
		"f4d24aa8fe81caab2420dcf4cf9ceb139394b535",
		"f9d9e55d1072d7a697d2bf06e1847e93635a7cf9",
		"df7dc4df69114c694956a0ac537119527ecf1b9c",
		"35e36c0cafbdc3395fd4600a05c613d3073c895e",
		"24d091b80d513846293c00350f46d85f71797aff",
		"f95b25589b40b5b0965deb592445073ff3efa299",
		"6a93ee522c52f5bd54140b6fa0be6a503e00dc96",
		"657d341c197d036dc27d7f8f5b61c6ff6a678df4",
		"cddd306a5010eba20a133b6473d9e8d967884f57",
		"0d825fd2e9e4dd42ac14d5ae6c7f92cbe63de009",
		"8bd7794fbdaa9536354dd2d8d961d9503beb9460",
		"e349501d4275363646a099e1d3baed064aa2eca5",
		"151dbcd21c9ed6b03960a5f0b05c255c9f955618",
		"28b0eaf7c500c506976da8d0fc9cad6c278e8d87",
		"a09a8c790a20309b942a9fdbfe77da22407096e6",
		"2c23f53ca22d7d8885fc4522ddcadcfe7f01a783",
		"65935d9855ece6f85c21ad38634703d0917bf88c",
		"dec00ef7c6155c4ca1109ec8248f7ff58d8f6cd3",
		"cc3d2b7b7cb6f077e3b1ee1d3e99eb54fddfa151",
		"009d724771e339ff7ec6cd7c0cc170d3470904c5",
		"e64aea8b539905fa92fad0e7cf73ffa4375f8b32",
		"6c62681a2f655b49963a5983b8b0950a6120ae14",
		"db708f7d959dee1857ac524636e85ecf2e1781c1",
		"2cd0a87ff7df953f810c344bdf2fe3340b954c69",
		"3aab2116756442bf0d4cd1c089b24d34c3baa253",
		"3af797a25458550a16bf89acc8e4ab2b7f2bfce0",
		"235a7e571b33eda1a81e0f73a3173ef95dd020e5",
		"50d0390056017158bdc75c063efd5c2a898d5f0c",
		"4205e3cf9c44264731ad002fcd2520eb1b2bb801",
		"53fc648efc0c82b1e0cc806ab7abf7dbdf532273",
		"faa8ba85d503da4ab872d17ba8c00da0098ab2f2",
		"7687a145717677e64300adeb44ac29d90f844b59",
		"814ec05f3683b661166055a23e29dca0300cd58c",
		"351719631846db88eb3daf690fca5399aed3fd77",
		"49c100caf72d658aca8e58bd74a4ba90fa2b0d70",
		"8cc35f73da321c29973191f2cf143d29d26a1df7",
		"a3f7325c52240418c2ba257b103c3c550e140c83",
		"7bb85ce2cf23af5b2d7467c2825fa2d0330ec5d5",
		"6fe2e3bb57daebd1555494818909f9664376dd6c",
		"1c63879e1f630e44ad8d2245b8a28a088f387e7c",
		"313913e603eaf3bb2c3b05079046ec07bb61f8c6",
		"4f062ad1aebb1255b84c851d00694cc7949de832",
		"887697058d8464462e8fd6d23c8461e90aec8c08",
		"2ab94758b0276a8a26102adf8d528cf6d0567b9a",
		"5397c9a02f77744da25d4ef63a7ebf01affeca62",
		"d6adb54fefe72482ed049f07af31ddf2c287345f",
		"4c65b7b43f3fe31350f74cb7d0b2461e111e8dd0",
		"e6feb6b7c06600924e8b6bda3263c870cfb0a447",
		"a09d2c48d3285f206fafbffe0e50619284e92126",
		"925720c5d40c4ebf8601e06025e1402251ef71d2",
		"611b82d4c4b4f67cc3d83cf0697ec660fcee2fff",
		"dfd5101b17da36c32ae024b984e0b72712f01a35",
		"68f1af10052713fda01bfb1e5b831dcf6d826ab2",
		"e2133b723d0e42be74880d34de6bf6538ea7f915",
		"e40429d9dd849c5fe0bdf97062b1d9358d99826d",
		"0ac2d2817d649e3203a8f7c93e7c65be0ca9662e",
		"7ef25e94db74d85fa7e9271b064a3c7d9ef7add5",
		"3e05dcce371d3f672feba29f086ad78a93ae3996",
		"16b9f8ab972e67eb21872ea2c40046249d543989",
		"c47579857bbf12c85499f431d4ecf27d77976b7c",
		"1ea4bec1a921180164852c65006d928617bd2caf",
		"d3ebf0f291297649b4c8dc3ecc81d2eddedc100d",
		"0ddae73613ab823639de096c287ea6142749f340",
		"6638e37b887b5a279044afbdc9928e19f678eb2e",
		"de7b8a41bbe1ccdfc009de51fa6d160db3ca8025",
		"f52de0603f31798455e48bd90e10a8f888dd6d93",
	}

	for i := 0; i < 5; i++ {
		t.Logf("Iteration %d", i+1)

		for _, sha := range shas {
			pkg, err := ms.GetMavenPackageBySha(ctx, sha)

			if err != nil {
				if strings.Contains(err.Error(), "no artifact found") {
					t.Logf("failed to get package by sha: %v", err)
					continue
				} else {
					t.Fatalf("failed to get package by sha: %v", err)
				}
			}

			// log human readable timestamp
			ti := time.Now()
			t.Logf("Time: %s Success: %s:%s", ti.String(), pkg.Name, pkg.Version)
		}
	}
}
