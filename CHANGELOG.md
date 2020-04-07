# Changelog

## [1.1.7](https://github.com/checkr/flagr/tree/1.1.7) (2020-03-19)

[Full Changelog](https://github.com/checkr/flagr/compare/1.1.6...HEAD)

**Closed issues:**

- When flagr encounters database errors during the `GetFlag` func it returns 404 [\#317](https://github.com/checkr/flagr/issues/317)

**Merged pull requests:**

- Add 404 get flag test [\#337](https://github.com/checkr/flagr/pull/337) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add publish dockerhub action [\#336](https://github.com/checkr/flagr/pull/336) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Update UI version [\#335](https://github.com/checkr/flagr/pull/335) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add changelog [\#334](https://github.com/checkr/flagr/pull/334) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add better error checking around 404s and actual database errors [\#318](https://github.com/checkr/flagr/pull/318) ([Cull-Methi](https://github.com/Cull-Methi))

## [1.1.6](https://github.com/checkr/flagr/tree/1.1.6) (2020-03-17)

[Full Changelog](https://github.com/checkr/flagr/compare/1.1.5...1.1.6)

**Fixed bugs:**

- UI appears to delete the wrong constraint [\#290](https://github.com/checkr/flagr/issues/290)

**Closed issues:**

- Question: Flags Usage Stats [\#325](https://github.com/checkr/flagr/issues/325)
- \[Bug\] When attempting to save an invalid JSON blob for a variant attachment, flagr reports success [\#324](https://github.com/checkr/flagr/issues/324)
- \[question/feature request\] Allow slashes \(/\) in flag keys [\#315](https://github.com/checkr/flagr/issues/315)
- \[Development\]Met error after executing 'make all' [\#309](https://github.com/checkr/flagr/issues/309)
- make fails with dependency error [\#303](https://github.com/checkr/flagr/issues/303)
- Need Feeedback on a client acting as a local evaluator [\#298](https://github.com/checkr/flagr/issues/298)
- Question: Other ways of setting created\_by or updated\_by [\#297](https://github.com/checkr/flagr/issues/297)
- Using `getFlag` and `findFlags` in python and go clients returns non-matching objects [\#294](https://github.com/checkr/flagr/issues/294)
- Variant Distribution does not work [\#293](https://github.com/checkr/flagr/issues/293)
- \[feat\] Add Flagr version to UI [\#287](https://github.com/checkr/flagr/issues/287)
- \[feat\] allow creating a flag with a specific ID [\#286](https://github.com/checkr/flagr/issues/286)
- UI not using correct value for ID [\#278](https://github.com/checkr/flagr/issues/278)

**Merged pull requests:**

- Bump version to 1.1.6 [\#333](https://github.com/checkr/flagr/pull/333) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump to go 1.14 [\#332](https://github.com/checkr/flagr/pull/332) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump npm package and fix security deps [\#329](https://github.com/checkr/flagr/pull/329) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix golangci-lint for context string key unittest [\#328](https://github.com/checkr/flagr/pull/328) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix json attachment validation and minimist dependency [\#327](https://github.com/checkr/flagr/pull/327) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump npm packages [\#321](https://github.com/checkr/flagr/pull/321) ([zhouzhuojie](https://github.com/zhouzhuojie))
- \[Feat\] Added Flagr version to UI [\#319](https://github.com/checkr/flagr/pull/319) ([wesleimp](https://github.com/wesleimp))
- Allow slashes in the flag name regex [\#316](https://github.com/checkr/flagr/pull/316) ([Cull-Methi](https://github.com/Cull-Methi))
- Remove fmt.Println from Prometheus middleware [\#313](https://github.com/checkr/flagr/pull/313) ([gfloyd](https://github.com/gfloyd))
- Fix swagger version [\#311](https://github.com/checkr/flagr/pull/311) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add ability to enable newrelic's distributed tracing [\#305](https://github.com/checkr/flagr/pull/305) ([jaysonsantos](https://github.com/jaysonsantos))
- Fix golangci-lint deps [\#304](https://github.com/checkr/flagr/pull/304) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Update flagr\_use\_cases.md [\#302](https://github.com/checkr/flagr/pull/302) ([sircelsius](https://github.com/sircelsius))
- Setting created\_by/updated\_by via header [\#300](https://github.com/checkr/flagr/pull/300) ([pacoguzman](https://github.com/pacoguzman))
- feat: Put EvalFlag public [\#299](https://github.com/checkr/flagr/pull/299) ([tkanos](https://github.com/tkanos))
- Change Postgres connection string example [\#296](https://github.com/checkr/flagr/pull/296) ([iJackUA](https://github.com/iJackUA))
- Fix UI for constraint deletion [\#295](https://github.com/checkr/flagr/pull/295) ([zhouzhuojie](https://github.com/zhouzhuojie))

## [1.1.5](https://github.com/checkr/flagr/tree/1.1.5) (2019-08-30)

[Full Changelog](https://github.com/checkr/flagr/compare/1.1.4...1.1.5)

**Closed issues:**

- Attempting to `make all` leads to undefined members due to improper capitalization [\#283](https://github.com/checkr/flagr/issues/283)
- How to deploy and use eval-only instances? [\#279](https://github.com/checkr/flagr/issues/279)

**Merged pull requests:**

- Bump eslint [\#288](https://github.com/checkr/flagr/pull/288) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump go-swagger [\#284](https://github.com/checkr/flagr/pull/284) ([zhouzhuojie](https://github.com/zhouzhuojie))
- added support for validating HS512 JWT tokens [\#282](https://github.com/checkr/flagr/pull/282) ([tejash-jl](https://github.com/tejash-jl))
- Use alpine 3.10 with glibc and bump remote docker in ciecleci [\#281](https://github.com/checkr/flagr/pull/281) ([zhouzhuojie](https://github.com/zhouzhuojie))
- docs: update Dynamic Configuration section [\#280](https://github.com/checkr/flagr/pull/280) ([kgeorgiou](https://github.com/kgeorgiou))
- Bump npm package [\#277](https://github.com/checkr/flagr/pull/277) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump version and add publish\_to\_docker script [\#276](https://github.com/checkr/flagr/pull/276) ([zhouzhuojie](https://github.com/zhouzhuojie))

## [1.1.4](https://github.com/checkr/flagr/tree/1.1.4) (2019-07-05)

[Full Changelog](https://github.com/checkr/flagr/compare/1.1.3...1.1.4)

**Closed issues:**

- Valid number range for constrains comparison? [\#274](https://github.com/checkr/flagr/issues/274)
- Strange characters coming through Kinesis stream [\#267](https://github.com/checkr/flagr/issues/267)
- Upgrade to vue-cli 3 for flagr-ui [\#264](https://github.com/checkr/flagr/issues/264)
- Variant Attachment JSON Support [\#231](https://github.com/checkr/flagr/issues/231)

**Merged pull requests:**

- Fix float comparison [\#275](https://github.com/checkr/flagr/pull/275) ([zhouzhuojie](https://github.com/zhouzhuojie))
- bugfix: "You deleted flag undefined" [\#273](https://github.com/checkr/flagr/pull/273) ([yosyad](https://github.com/yosyad))
- Use lower case and computed for search terms [\#271](https://github.com/checkr/flagr/pull/271) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Height fix for json-editor in variant attachment [\#269](https://github.com/checkr/flagr/pull/269) ([yosyad](https://github.com/yosyad))
- Add search bar for filtering flags in home page [\#268](https://github.com/checkr/flagr/pull/268) ([yosyad](https://github.com/yosyad))

## [1.1.3](https://github.com/checkr/flagr/tree/1.1.3) (2019-05-30)

[Full Changelog](https://github.com/checkr/flagr/compare/1.1.2...1.1.3)

**Closed issues:**

- Bulk Evaluation Doesn't Return Variant Information [\#263](https://github.com/checkr/flagr/issues/263)
- Healthcheck Downloading Gzip file? [\#261](https://github.com/checkr/flagr/issues/261)

**Merged pull requests:**

- Fix assetsDir [\#266](https://github.com/checkr/flagr/pull/266) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Remove yarn and upgrade to vue-cli 3 [\#265](https://github.com/checkr/flagr/pull/265) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add support for arbitrary JSON in attachments [\#259](https://github.com/checkr/flagr/pull/259) ([saary](https://github.com/saary))
- Changed input text to json-editor in flag variant form [\#258](https://github.com/checkr/flagr/pull/258) ([yosyad](https://github.com/yosyad))

## [1.1.2](https://github.com/checkr/flagr/tree/1.1.2) (2019-05-24)

[Full Changelog](https://github.com/checkr/flagr/compare/1.1.1...1.1.2)

**Closed issues:**

- Arbitrary validation rules in feature and variant keys [\#254](https://github.com/checkr/flagr/issues/254)
- Release 1.1.1 is missing from Docker Hub [\#251](https://github.com/checkr/flagr/issues/251)
- Fatal Error when Unable to Reach Kafka [\#244](https://github.com/checkr/flagr/issues/244)
- Support disk files as data source and make flagr read-only under that mode [\#237](https://github.com/checkr/flagr/issues/237)
- Release Schedule [\#211](https://github.com/checkr/flagr/issues/211)
- Support Go Module [\#201](https://github.com/checkr/flagr/issues/201)

**Merged pull requests:**

- Fix health check endpoint [\#262](https://github.com/checkr/flagr/pull/262) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump webpack-bundle-analyzer [\#260](https://github.com/checkr/flagr/pull/260) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Allow more capital characters and numbers in flag keys and variant keys [\#257](https://github.com/checkr/flagr/pull/257) ([raviambati](https://github.com/raviambati))
- Add ability to output logs in JSON format [\#253](https://github.com/checkr/flagr/pull/253) ([croemmich](https://github.com/croemmich))
- Pin flagr-ci docker version tag [\#250](https://github.com/checkr/flagr/pull/250) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix lint and make vendor [\#249](https://github.com/checkr/flagr/pull/249) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add go module support [\#248](https://github.com/checkr/flagr/pull/248) ([zhouzhuojie](https://github.com/zhouzhuojie))

## [1.1.1](https://github.com/checkr/flagr/tree/1.1.1) (2019-04-16)

[Full Changelog](https://github.com/checkr/flagr/compare/1.1.0...1.1.1)

**Implemented enhancements:**

- Flag Note Support [\#230](https://github.com/checkr/flagr/issues/230)

**Closed issues:**

- WebPrefix does not update API\_URL [\#225](https://github.com/checkr/flagr/issues/225)

**Merged pull requests:**

- Extend DBDRIVER to load from json\_file or json\_http and add EvalOnlyMode [\#247](https://github.com/checkr/flagr/pull/247) ([zhouzhuojie](https://github.com/zhouzhuojie))
- ensure header color css applies only within el-card\_\_header [\#246](https://github.com/checkr/flagr/pull/246) ([crberube](https://github.com/crberube))
- Call data recorder at the startup time [\#245](https://github.com/checkr/flagr/pull/245) ([zhouzhuojie](https://github.com/zhouzhuojie))
- add note support to flags [\#243](https://github.com/checkr/flagr/pull/243) ([crberube](https://github.com/crberube))
- Remove entity\_type statsd tag [\#242](https://github.com/checkr/flagr/pull/242) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Remove required fields from evalResult swagger model [\#239](https://github.com/checkr/flagr/pull/239) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Use web prefix for main API [\#234](https://github.com/checkr/flagr/pull/234) ([zhouzhuojie](https://github.com/zhouzhuojie))

## [1.1.0](https://github.com/checkr/flagr/tree/1.1.0) (2019-03-19)

[Full Changelog](https://github.com/checkr/flagr/compare/1.0.15...1.1.0)

**Implemented enhancements:**

- Consolidate the evalResult for kafka, kinesis and pubsub logging [\#203](https://github.com/checkr/flagr/issues/203)

## [1.0.15](https://github.com/checkr/flagr/tree/1.0.15) (2019-03-11)

[Full Changelog](https://github.com/checkr/flagr/compare/1.0.14...1.0.15)

**Closed issues:**

- Data Pipeline Format [\#232](https://github.com/checkr/flagr/issues/232)
- Disable enableDebug Propery for Evaluation Transactions [\#228](https://github.com/checkr/flagr/issues/228)
- Flagr-UI Cannot Resolve Static Assets when using FLAGR\_WEB\_PREFIX environment variable [\#222](https://github.com/checkr/flagr/issues/222)

**Merged pull requests:**

- Add EvalDebugEnabled env variable [\#235](https://github.com/checkr/flagr/pull/235) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Refactor data recorder [\#233](https://github.com/checkr/flagr/pull/233) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add release badge [\#227](https://github.com/checkr/flagr/pull/227) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump version [\#224](https://github.com/checkr/flagr/pull/224) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump conditions [\#219](https://github.com/checkr/flagr/pull/219) ([zhouzhuojie](https://github.com/zhouzhuojie))

## [1.0.14](https://github.com/checkr/flagr/tree/1.0.14) (2019-02-22)

[Full Changelog](https://github.com/checkr/flagr/compare/1.0.13...1.0.14)

**Implemented enhancements:**

- Support Prometheus [\#196](https://github.com/checkr/flagr/issues/196)

**Closed issues:**

- Swagger Codegen generated PHP SDK [\#220](https://github.com/checkr/flagr/issues/220)
- Retry DB Connections [\#207](https://github.com/checkr/flagr/issues/207)

**Merged pull requests:**

- Use English as default locale for flagr instead of Chinese [\#226](https://github.com/checkr/flagr/pull/226) ([lawrenceong](https://github.com/lawrenceong))
- Use relative assetsPublicPath [\#223](https://github.com/checkr/flagr/pull/223) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Adds Prometheus export \(requests and variant eval\) [\#221](https://github.com/checkr/flagr/pull/221) ([jasongwartz](https://github.com/jasongwartz))
- Revert "Bump vendors" [\#218](https://github.com/checkr/flagr/pull/218) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump vendors [\#217](https://github.com/checkr/flagr/pull/217) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix empty response when flag was deleted [\#213](https://github.com/checkr/flagr/pull/213) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Added Google Cloud Pubsub support for data records [\#209](https://github.com/checkr/flagr/pull/209) ([vic3lord](https://github.com/vic3lord))

## [1.0.13](https://github.com/checkr/flagr/tree/1.0.13) (2019-01-30)

[Full Changelog](https://github.com/checkr/flagr/compare/1.0.12...1.0.13)

**Closed issues:**

- Empty Arrays Returned for Segments and Variants on GET flags endpoint [\#199](https://github.com/checkr/flagr/issues/199)
- Granular access control for Flagr UI [\#195](https://github.com/checkr/flagr/issues/195)
- Getting path / was not found on new install [\#192](https://github.com/checkr/flagr/issues/192)
- findFlags not working with `key` query param [\#187](https://github.com/checkr/flagr/issues/187)
- Flagr segments evaluation should stop if it matches all the constraints in a segment [\#180](https://github.com/checkr/flagr/issues/180)
- Question: Using the JWT Auth [\#121](https://github.com/checkr/flagr/issues/121)

**Merged pull requests:**

- Add retry for DB connection [\#208](https://github.com/checkr/flagr/pull/208) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix golint [\#206](https://github.com/checkr/flagr/pull/206) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add datadog apm support [\#205](https://github.com/checkr/flagr/pull/205) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Replace '+' icon with 'Add Constraint' to make usage clearer [\#202](https://github.com/checkr/flagr/pull/202) ([erdey](https://github.com/erdey))
- Add preload param in get /flags [\#200](https://github.com/checkr/flagr/pull/200) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Update README.md [\#198](https://github.com/checkr/flagr/pull/198) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Improve docs [\#194](https://github.com/checkr/flagr/pull/194) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix Makefile and docs [\#193](https://github.com/checkr/flagr/pull/193) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Reorder middlewares and add test for must parse kafka version [\#191](https://github.com/checkr/flagr/pull/191) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add integration tests [\#190](https://github.com/checkr/flagr/pull/190) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add kafka version config [\#189](https://github.com/checkr/flagr/pull/189) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Remove goqueryset and use gorm directly [\#188](https://github.com/checkr/flagr/pull/188) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add headers to CORS middleware [\#186](https://github.com/checkr/flagr/pull/186) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Make user claim configurable [\#183](https://github.com/checkr/flagr/pull/183) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Update version and docs [\#182](https://github.com/checkr/flagr/pull/182) ([zhouzhuojie](https://github.com/zhouzhuojie))

## [1.0.12](https://github.com/checkr/flagr/tree/1.0.12) (2018-10-23)

[Full Changelog](https://github.com/checkr/flagr/compare/1.0.11...1.0.12)

**Merged pull requests:**

- Fix segments evaluation [\#181](https://github.com/checkr/flagr/pull/181) ([zhouzhuojie](https://github.com/zhouzhuojie))

## [1.0.11](https://github.com/checkr/flagr/tree/1.0.11) (2018-10-10)

[Full Changelog](https://github.com/checkr/flagr/compare/1.0.10...1.0.11)

**Closed issues:**

- Support for liveness check in flagr [\#165](https://github.com/checkr/flagr/issues/165)
- Save Segment button is not clear if it saves the constraints or not [\#142](https://github.com/checkr/flagr/issues/142)

**Merged pull requests:**

- Remove some dd metrics [\#179](https://github.com/checkr/flagr/pull/179) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add kafka data recorder dd metrics [\#178](https://github.com/checkr/flagr/pull/178) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add dd metrics [\#177](https://github.com/checkr/flagr/pull/177) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix mysql query [\#176](https://github.com/checkr/flagr/pull/176) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add UI env for entity types override [\#175](https://github.com/checkr/flagr/pull/175) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix demo db and bump ui [\#173](https://github.com/checkr/flagr/pull/173) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add entity\_type override [\#171](https://github.com/checkr/flagr/pull/171) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Added config for base path for web UI [\#170](https://github.com/checkr/flagr/pull/170) ([SebastianOsuna](https://github.com/SebastianOsuna))
- Fix operator width UI [\#169](https://github.com/checkr/flagr/pull/169) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Use pointer cursor for flags table rows [\#168](https://github.com/checkr/flagr/pull/168) ([marceloboeira](https://github.com/marceloboeira))
- Add gzip middleware [\#167](https://github.com/checkr/flagr/pull/167) ([zhouzhuojie](https://github.com/zhouzhuojie))

## [1.0.10](https://github.com/checkr/flagr/tree/1.0.10) (2018-09-17)

[Full Changelog](https://github.com/checkr/flagr/compare/1.0.9...1.0.10)

**Closed issues:**

- Question - How does a change in distribution affect existing users? [\#161](https://github.com/checkr/flagr/issues/161)

**Merged pull requests:**

- Improve homepage UI [\#164](https://github.com/checkr/flagr/pull/164) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix 401 jwt middleware handling [\#163](https://github.com/checkr/flagr/pull/163) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Improve UI [\#162](https://github.com/checkr/flagr/pull/162) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add flagr key [\#159](https://github.com/checkr/flagr/pull/159) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump vendor [\#158](https://github.com/checkr/flagr/pull/158) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add helper script for getting remote sqlite [\#157](https://github.com/checkr/flagr/pull/157) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add curl in dockerfile [\#156](https://github.com/checkr/flagr/pull/156) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Update docs and add operation id [\#155](https://github.com/checkr/flagr/pull/155) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add export to sqlite feature [\#154](https://github.com/checkr/flagr/pull/154) ([zhouzhuojie](https://github.com/zhouzhuojie))

## [1.0.9](https://github.com/checkr/flagr/tree/1.0.9) (2018-08-24)

[Full Changelog](https://github.com/checkr/flagr/compare/1.0.8...1.0.9)

**Closed issues:**

- Kinesis Data Recorder Adapter [\#150](https://github.com/checkr/flagr/issues/150)

**Merged pull requests:**

- Change random key generation and db logging for migration [\#160](https://github.com/checkr/flagr/pull/160) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump vendors [\#153](https://github.com/checkr/flagr/pull/153) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix typo on constraints error alert [\#152](https://github.com/checkr/flagr/pull/152) ([chrisivens](https://github.com/chrisivens))
- Add Kinesis support [\#151](https://github.com/checkr/flagr/pull/151) ([marceloboeira](https://github.com/marceloboeira))
- Change rebuild Makefile cmd [\#147](https://github.com/checkr/flagr/pull/147) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add JWT auth via headers and RS256 signing option [\#146](https://github.com/checkr/flagr/pull/146) ([vayan](https://github.com/vayan))
- Add coverage for kafka data recorder [\#145](https://github.com/checkr/flagr/pull/145) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add test coverage for middleware [\#144](https://github.com/checkr/flagr/pull/144) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump UI packages [\#143](https://github.com/checkr/flagr/pull/143) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix docs [\#141](https://github.com/checkr/flagr/pull/141) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add tests for find flags parameters [\#138](https://github.com/checkr/flagr/pull/138) ([zhouzhuojie](https://github.com/zhouzhuojie))

## [1.0.8](https://github.com/checkr/flagr/tree/1.0.8) (2018-07-06)

[Full Changelog](https://github.com/checkr/flagr/compare/1.0.7...1.0.8)

**Closed issues:**

- Deploying Flagr Architecture [\#139](https://github.com/checkr/flagr/issues/139)
- getting error when running make build [\#133](https://github.com/checkr/flagr/issues/133)
- Java SDK [\#128](https://github.com/checkr/flagr/issues/128)
- Monitoring / Heath Check  [\#102](https://github.com/checkr/flagr/issues/102)
- Roadmap [\#100](https://github.com/checkr/flagr/issues/100)

**Merged pull requests:**

- Add health handler [\#140](https://github.com/checkr/flagr/pull/140) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix InsecureVerifySSL config [\#137](https://github.com/checkr/flagr/pull/137) ([zhouzhuojie](https://github.com/zhouzhuojie))

## [1.0.7](https://github.com/checkr/flagr/tree/1.0.7) (2018-06-26)

[Full Changelog](https://github.com/checkr/flagr/compare/1.0.6...1.0.7)

**Closed issues:**

- Query params for findFlags not working [\#132](https://github.com/checkr/flagr/issues/132)

**Merged pull requests:**

- Fix sed so it works for both mac and gnu sed [\#136](https://github.com/checkr/flagr/pull/136) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Lock dev tools with retool [\#135](https://github.com/checkr/flagr/pull/135) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add offset support [\#134](https://github.com/checkr/flagr/pull/134) ([zhouzhuojie](https://github.com/zhouzhuojie))

## [1.0.6](https://github.com/checkr/flagr/tree/1.0.6) (2018-06-18)

[Full Changelog](https://github.com/checkr/flagr/compare/1.0.5...1.0.6)

**Merged pull requests:**

- Add description\_like query param [\#131](https://github.com/checkr/flagr/pull/131) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump vendor and gen tools [\#130](https://github.com/checkr/flagr/pull/130) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Remove default limit for find flags [\#129](https://github.com/checkr/flagr/pull/129) ([zhouzhuojie](https://github.com/zhouzhuojie))
- add support for querying flags by description, enabled and with limit [\#126](https://github.com/checkr/flagr/pull/126) ([amalfra](https://github.com/amalfra))
- Bump vendor [\#125](https://github.com/checkr/flagr/pull/125) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump deps [\#124](https://github.com/checkr/flagr/pull/124) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix typo in flagr overview docs [\#123](https://github.com/checkr/flagr/pull/123) ([markphelps](https://github.com/markphelps))
- Add db connection debugging logging env [\#122](https://github.com/checkr/flagr/pull/122) ([zhouzhuojie](https://github.com/zhouzhuojie))

## [1.0.5](https://github.com/checkr/flagr/tree/1.0.5) (2018-05-01)

[Full Changelog](https://github.com/checkr/flagr/compare/1.0.4...1.0.5)

**Closed issues:**

- Respect log level [\#115](https://github.com/checkr/flagr/issues/115)

**Merged pull requests:**

- Bump UI vendor [\#120](https://github.com/checkr/flagr/pull/120) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump conditions to support null json value [\#119](https://github.com/checkr/flagr/pull/119) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Use random entity id if it's nil [\#118](https://github.com/checkr/flagr/pull/118) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Use env to config eval logging [\#117](https://github.com/checkr/flagr/pull/117) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix test warning [\#116](https://github.com/checkr/flagr/pull/116) ([zhouzhuojie](https://github.com/zhouzhuojie))

## [1.0.4](https://github.com/checkr/flagr/tree/1.0.4) (2018-04-12)

[Full Changelog](https://github.com/checkr/flagr/compare/1.0.3...1.0.4)

**Closed issues:**

- pkg/repo/db.go not throwing fatal erros when issue in connecting DB [\#110](https://github.com/checkr/flagr/issues/110)
- Cannot build from source on macOS [\#106](https://github.com/checkr/flagr/issues/106)
- Using Alpine 3.6 version as base image [\#105](https://github.com/checkr/flagr/issues/105)
- Can't create new flags with pgsql [\#101](https://github.com/checkr/flagr/issues/101)

**Merged pull requests:**

- Use embeded file for env docs [\#114](https://github.com/checkr/flagr/pull/114) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix postgres string scan bug [\#113](https://github.com/checkr/flagr/pull/113) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Push http metrics with statsd [\#112](https://github.com/checkr/flagr/pull/112) ([marceloboeira](https://github.com/marceloboeira))
- Use fatal instead of panic for db connection problem [\#111](https://github.com/checkr/flagr/pull/111) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump newrelic go-agent and add ca-certificates apk [\#109](https://github.com/checkr/flagr/pull/109) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump vendor [\#108](https://github.com/checkr/flagr/pull/108) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Use multi-stage build for dockerfile [\#107](https://github.com/checkr/flagr/pull/107) ([zhouzhuojie](https://github.com/zhouzhuojie))

## [1.0.3](https://github.com/checkr/flagr/tree/1.0.3) (2018-03-19)

[Full Changelog](https://github.com/checkr/flagr/compare/1.0.2...1.0.3)

**Merged pull requests:**

- Fix constraints composition bug [\#104](https://github.com/checkr/flagr/pull/104) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add test coverage in pkg/handler [\#103](https://github.com/checkr/flagr/pull/103) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add console logging rate limiter [\#99](https://github.com/checkr/flagr/pull/99) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add an option for middleware verbose logging [\#98](https://github.com/checkr/flagr/pull/98) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add cast package for safe cast [\#97](https://github.com/checkr/flagr/pull/97) ([zhouzhuojie](https://github.com/zhouzhuojie))

## [1.0.2](https://github.com/checkr/flagr/tree/1.0.2) (2018-01-16)

[Full Changelog](https://github.com/checkr/flagr/compare/1.0.1...1.0.2)

## [1.0.1](https://github.com/checkr/flagr/tree/1.0.1) (2018-01-16)

[Full Changelog](https://github.com/checkr/flagr/compare/1.0.0...1.0.1)

## [1.0.0](https://github.com/checkr/flagr/tree/1.0.0) (2018-01-16)

[Full Changelog](https://github.com/checkr/flagr/compare/a012393f0d70f1d1d84b439e7b222e4e980cee52...1.0.0)

**Merged pull requests:**

- Add range in codecov.yml [\#96](https://github.com/checkr/flagr/pull/96) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Make SaveFlagSnapshot sync and increase coverage [\#95](https://github.com/checkr/flagr/pull/95) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Increase coverage [\#94](https://github.com/checkr/flagr/pull/94) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add crud\_test unit tests coverage [\#93](https://github.com/checkr/flagr/pull/93) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump vendor and change coverage tool [\#92](https://github.com/checkr/flagr/pull/92) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Update codecov.yml to actually ignore paths [\#91](https://github.com/checkr/flagr/pull/91) ([kruppel](https://github.com/kruppel))
- Add eager preload to avoid n+1 problem [\#90](https://github.com/checkr/flagr/pull/90) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix the preload logic for segments [\#89](https://github.com/checkr/flagr/pull/89) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix prefix whitelist and no changes snapshots [\#88](https://github.com/checkr/flagr/pull/88) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add UI to show the history [\#87](https://github.com/checkr/flagr/pull/87) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add snapshots endpoint [\#86](https://github.com/checkr/flagr/pull/86) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Update dockerignore [\#85](https://github.com/checkr/flagr/pull/85) ([zhouzhuojie](https://github.com/zhouzhuojie))
- UI fix and wording [\#84](https://github.com/checkr/flagr/pull/84) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix flag snapshot ID in eval [\#83](https://github.com/checkr/flagr/pull/83) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Save updated\_by snapshot of flags [\#82](https://github.com/checkr/flagr/pull/82) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump deps [\#81](https://github.com/checkr/flagr/pull/81) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add JWT auth login [\#80](https://github.com/checkr/flagr/pull/80) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix api docs title [\#79](https://github.com/checkr/flagr/pull/79) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Create CODE\_OF\_CONDUCT.md [\#78](https://github.com/checkr/flagr/pull/78) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Update README [\#77](https://github.com/checkr/flagr/pull/77) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix relative URLs [\#76](https://github.com/checkr/flagr/pull/76) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Update docs, docker images, and configuration [\#75](https://github.com/checkr/flagr/pull/75) ([zhouzhuojie](https://github.com/zhouzhuojie))
- UI fixes [\#74](https://github.com/checkr/flagr/pull/74) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix codecov [\#73](https://github.com/checkr/flagr/pull/73) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Improve the code coverage [\#72](https://github.com/checkr/flagr/pull/72) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add more docs [\#71](https://github.com/checkr/flagr/pull/71) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add timeout to graceful server [\#70](https://github.com/checkr/flagr/pull/70) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add sleep [\#69](https://github.com/checkr/flagr/pull/69) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add coverage report [\#68](https://github.com/checkr/flagr/pull/68) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add go-report card [\#67](https://github.com/checkr/flagr/pull/67) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix debug console layout problem [\#66](https://github.com/checkr/flagr/pull/66) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix segment stable sort with asc segmentID [\#65](https://github.com/checkr/flagr/pull/65) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix distribution checkbox [\#64](https://github.com/checkr/flagr/pull/64) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Fix new relic client nil pointer error [\#63](https://github.com/checkr/flagr/pull/63) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add sentry and newrelic monitoring [\#62](https://github.com/checkr/flagr/pull/62) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add debug console [\#61](https://github.com/checkr/flagr/pull/61) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Change back to string format and only stdout with evalResult [\#60](https://github.com/checkr/flagr/pull/60) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Bump to element-ui 2.0 [\#59](https://github.com/checkr/flagr/pull/59) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add segments reorder [\#58](https://github.com/checkr/flagr/pull/58) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Support constraint editing - putConstraint [\#57](https://github.com/checkr/flagr/pull/57) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add api docs button and data records toggle [\#56](https://github.com/checkr/flagr/pull/56) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add JSON indent [\#55](https://github.com/checkr/flagr/pull/55) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Make data records optional with the kafka message frame schema [\#54](https://github.com/checkr/flagr/pull/54) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add favicon! [\#53](https://github.com/checkr/flagr/pull/53) ([lucidrains](https://github.com/lucidrains))
- Add new relic monitoring [\#52](https://github.com/checkr/flagr/pull/52) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Truncate constraint value and add a tooltip [\#51](https://github.com/checkr/flagr/pull/51) ([lucidrains](https://github.com/lucidrains))
- Add ability to delete variant. Halt user if the variant is currently … [\#50](https://github.com/checkr/flagr/pull/50) ([lucidrains](https://github.com/lucidrains))
- Fix logrus and unify all the stdout logging [\#49](https://github.com/checkr/flagr/pull/49) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add batch evaluation [\#48](https://github.com/checkr/flagr/pull/48) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add eval logs to kibana [\#47](https://github.com/checkr/flagr/pull/47) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add edit segment button [\#46](https://github.com/checkr/flagr/pull/46) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add ability to delete segment [\#45](https://github.com/checkr/flagr/pull/45) ([lucidrains](https://github.com/lucidrains))
- Change env lib [\#44](https://github.com/checkr/flagr/pull/44) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add variant key validation and put it in the evaluation result [\#43](https://github.com/checkr/flagr/pull/43) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Change variant key should also change distribution [\#42](https://github.com/checkr/flagr/pull/42) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Editing variants [\#41](https://github.com/checkr/flagr/pull/41) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add dockerfile and new introduction doc [\#40](https://github.com/checkr/flagr/pull/40) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add docs [\#39](https://github.com/checkr/flagr/pull/39) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add put and delete for segments [\#38](https://github.com/checkr/flagr/pull/38) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add put and delete for variants [\#37](https://github.com/checkr/flagr/pull/37) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add kafka data logging [\#36](https://github.com/checkr/flagr/pull/36) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Combine the constraints evaluation logic [\#35](https://github.com/checkr/flagr/pull/35) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Some UI tweaks [\#34](https://github.com/checkr/flagr/pull/34) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add constraints validation [\#33](https://github.com/checkr/flagr/pull/33) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add ability to delete constraints and also add indicator on flags col… [\#32](https://github.com/checkr/flagr/pull/32) ([lucidrains](https://github.com/lucidrains))
- Put and Delete constraints [\#31](https://github.com/checkr/flagr/pull/31) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add watcher [\#30](https://github.com/checkr/flagr/pull/30) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add ability to enable / disable feature flag [\#29](https://github.com/checkr/flagr/pull/29) ([lucidrains](https://github.com/lucidrains))
- Break the huge swagger file into files [\#28](https://github.com/checkr/flagr/pull/28) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Strong validation on putDistribution [\#27](https://github.com/checkr/flagr/pull/27) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add pprof [\#26](https://github.com/checkr/flagr/pull/26) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add enabled field for flag [\#25](https://github.com/checkr/flagr/pull/25) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add attachment for variant [\#24](https://github.com/checkr/flagr/pull/24) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add evaluator implementation [\#23](https://github.com/checkr/flagr/pull/23) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Update UI README [\#22](https://github.com/checkr/flagr/pull/22) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Move form for creating segment into dialog [\#21](https://github.com/checkr/flagr/pull/21) ([lucidrains](https://github.com/lucidrains))
- Change Makefile [\#20](https://github.com/checkr/flagr/pull/20) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Do not allow distribution saving unless percentages add up to 100%. A… [\#19](https://github.com/checkr/flagr/pull/19) ([lucidrains](https://github.com/lucidrains))
- Add evaluation cache [\#18](https://github.com/checkr/flagr/pull/18) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Prepare flag evaluation [\#17](https://github.com/checkr/flagr/pull/17) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add an error message if initial flag list cannot be loaded [\#16](https://github.com/checkr/flagr/pull/16) ([lucidrains](https://github.com/lucidrains))
- Preload flag for get flag [\#15](https://github.com/checkr/flagr/pull/15) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add ability to create and display variants per feature flag in UI [\#14](https://github.com/checkr/flagr/pull/14) ([lucidrains](https://github.com/lucidrains))
- Add element-ui library, carve out interface for adding segments and c… [\#13](https://github.com/checkr/flagr/pull/13) ([lucidrains](https://github.com/lucidrains))
- Add crud variants [\#12](https://github.com/checkr/flagr/pull/12) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Remove rank in distribution model [\#11](https://github.com/checkr/flagr/pull/11) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add CRUD distribution and fix segment create [\#10](https://github.com/checkr/flagr/pull/10) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add basic frontend for displaying list of feature flags on front page… [\#9](https://github.com/checkr/flagr/pull/9) ([lucidrains](https://github.com/lucidrains))
- Add constraints route [\#8](https://github.com/checkr/flagr/pull/8) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add segment create and read [\#7](https://github.com/checkr/flagr/pull/7) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add put and delete flag route [\#6](https://github.com/checkr/flagr/pull/6) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add GetFlag route [\#5](https://github.com/checkr/flagr/pull/5) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Minimal FindFlags api [\#4](https://github.com/checkr/flagr/pull/4) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add static file route [\#3](https://github.com/checkr/flagr/pull/3) ([zhouzhuojie](https://github.com/zhouzhuojie))
- Add build scripts to flagr-ui project so dev environment can work [\#2](https://github.com/checkr/flagr/pull/2) ([lucidrains](https://github.com/lucidrains))
- Update README [\#1](https://github.com/checkr/flagr/pull/1) ([zhouzhuojie](https://github.com/zhouzhuojie))



\* *This Changelog was automatically generated by [github_changelog_generator](https://github.com/github-changelog-generator/github-changelog-generator)*
