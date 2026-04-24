import type { CompilerError, SFCDescriptor } from '@vue/compiler-sfc';
export declare function initVueTemplateCompilerExtraOptions(descriptor: SFCDescriptor): {
    root: string;
    platform: keyof UniApp.PagesJsonPagePlatformStyle;
    componentType: string;
    filename: string;
    relativeFilename: string;
    helper: any;
    scriptCppBlocks: any;
    disableStaticStyle: boolean;
    onVueTemplateCompileLog(type: 'warn' | 'error', error: CompilerError): void;
    r: any;
    className: any;
    inlineRender: boolean;
};
