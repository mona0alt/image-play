"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.initVueTemplateCompilerExtraOptions = void 0;
const path_1 = __importDefault(require("path"));
const utils_1 = require("../utils");
const json_1 = require("../json");
const vue_1 = require("../vue");
function initVueTemplateCompilerExtraOptions(descriptor) {
    const filename = (0, utils_1.normalizePath)(descriptor.filename.split('?')[0]);
    const relativeFilename = (0, utils_1.normalizePath)(path_1.default.relative(process.env.UNI_INPUT_DIR, filename));
    const isDevX = process.env.UNI_HX_VERSION_DEV === 'true' &&
        process.env.UNI_APP_X === 'true';
    let disableStaticStyle = false;
    if (isDevX && process.env.NODE_ENV === 'development') {
        if (process.env.UNI_UTS_PLATFORM === 'app-harmony') {
            // 开发版本、开发模式下，非鸿蒙release模式打包
            disableStaticStyle = process.env.UNI_APP_HARMONY_RUN_MODE !== 'release';
        }
    }
    const helper = (0, utils_1.requireUniHelpers)();
    return {
        root: (0, utils_1.normalizePath)(process.env.UNI_INPUT_DIR),
        platform: process.env.UNI_UTS_PLATFORM,
        componentType: (0, json_1.isUniPageFile)(filename) ? 'page' : 'component',
        filename: filename,
        relativeFilename,
        helper,
        scriptCppBlocks: descriptor.scriptCppBlocks,
        disableStaticStyle,
        onVueTemplateCompileLog(type, error) {
            return (0, vue_1.onVueTemplateCompileLog)(type, error, descriptor.source, relativeFilename);
        },
        r: helper.K,
        className: helper.GCN(descriptor.filename, process.env.UNI_INPUT_DIR),
        inlineRender: process.env.UNI_UTS_PLATFORM === 'app-android',
    };
}
exports.initVueTemplateCompilerExtraOptions = initVueTemplateCompilerExtraOptions;
