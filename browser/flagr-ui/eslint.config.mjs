import pluginVue from 'eslint-plugin-vue'
import vueParser from 'vue-eslint-parser'
import babelParser from '@babel/eslint-parser'

export default [
  ...pluginVue.configs['flat/essential'],
  {
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: babelParser,
        requireConfigFile: false,
        ecmaVersion: 'latest',
        sourceType: 'module',
      },
    },
    rules: {
      'vue/multi-word-component-names': 'off',
      'vue/no-reserved-component-names': 'off',
    },
  },
]
