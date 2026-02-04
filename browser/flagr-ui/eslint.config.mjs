import js from '@eslint/js';
import pluginVue from 'eslint-plugin-vue';
import babelParser from '@babel/eslint-parser';
import globals from 'globals';

export default [
  {
    ignores: ['dist/', 'node_modules/'],
  },

  js.configs.recommended,

  ...pluginVue.configs['flat/recommended'],

  {
    files: ['src/**/*.{js,vue}'],
    languageOptions: {
      parserOptions: {
        parser: babelParser,
        requireConfigFile: false,
      },
      globals: {
        ...globals.browser,
        process: 'readonly',
      },
    },
    rules: {
      'no-console': 'warn',
      'vue/multi-word-component-names': 'off',
      'vue/no-v-html': 'off',
      'vue/no-ref-object-reactivity-loss': 'error',
      'vue/define-macros-order': ['warn', {
        order: ['defineProps', 'defineEmits'],
      }],
    },
  },
];
