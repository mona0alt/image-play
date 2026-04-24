"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.uniAppXAndroidEnginePlugin = void 0;
const uni_cli_shared_1 = require("@dcloudio/uni-cli-shared");
function uniAppXAndroidEnginePlugin() {
    const { compileVaporApp, getKotlinCompilerServer } = (0, uni_cli_shared_1.resolveUTSCompiler)();
    const compilerServer = getKotlinCompilerServer();
    if (!compilerServer) {
        throw new Error(`项目使用了uts插件，正在安装 uts Android 运行扩展...`);
    }
    const outputDir = process.env.UNI_OUTPUT_DIR;
    const uvueOutputDir = (0, uni_cli_shared_1.uvueOutDir)('app-android');
    const { UKF } = (0, uni_cli_shared_1.requireUniHelpers)();
    return {
        name: 'uni:app-x-android',
        async writeBundle() {
            if (!compilerServer) {
                return;
            }
            if (process.env.UNI_APP_X_DOM2_CPP_CHANGED === 'true' ||
                process.env.UNI_APP_X_DOM2_KT_CHANGED === 'true') {
                const { changed, files } = UKF();
                await compileVaporApp({
                    filename: 'index.kt',
                    changed: changed,
                    chunks: files,
                    inputDir: uvueOutputDir,
                    outputDir: outputDir,
                });
            }
        },
    };
}
exports.uniAppXAndroidEnginePlugin = uniAppXAndroidEnginePlugin;
