// eslint.config.js
import path from "node:path";
import { fileURLToPath } from "node:url";

import js from "@eslint/js";
import { defineConfig } from "eslint/config";
import globals from "globals";
import tseslint from "typescript-eslint";

import { createTypeScriptImportResolver } from "eslint-import-resolver-typescript";
import { createNodeResolver, importX } from "eslint-plugin-import-x";
import reactHooks from "eslint-plugin-react-hooks";
import sonarjs from "eslint-plugin-sonarjs";
import unicorn from "eslint-plugin-unicorn";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig(
  { ignores: ["dist", "build", "coverage", "node_modules", "src/api/openapi-generated"] },

  { files: ["**/*.{ts,tsx}"] },

  js.configs.recommended,

  ...tseslint.configs.strictTypeChecked,
  ...tseslint.configs.stylisticTypeChecked,

  importX.flatConfigs.recommended,
  importX.flatConfigs.typescript,

  sonarjs.configs.recommended,

  unicorn.configs.recommended,

  {
    languageOptions: {
      ecmaVersion: "latest",
      sourceType: "module",
      globals: {
        ...globals.browser,
        ...globals.es2021,
      },
      parserOptions: {
        tsconfigRootDir: __dirname,
        projectService: true,
      },
    },
    settings: {
      "import-x/resolver-next": [
        createTypeScriptImportResolver({ alwaysTryTypes: true, project: "./tsconfig.json" }),
        createNodeResolver(),
      ],
    },
    plugins: {
      "react-hooks": reactHooks,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,

      "no-console": "error",
      "no-debugger": "error",
      "import-x/no-unresolved": "off",

      "@typescript-eslint/consistent-type-imports": ["error", { fixStyle: "inline-type-imports" }],
      "@typescript-eslint/no-floating-promises": ["error", { ignoreVoid: false, ignoreIIFE: false }],
      "@typescript-eslint/no-misused-promises": ["error", { checksVoidReturn: true }],
      "@typescript-eslint/explicit-function-return-type": [
        "error",
        { allowExpressions: false, allowTypedFunctionExpressions: false },
      ],
      "unicorn/filename-case": "off",
    },
  },
  {
    files: ["src/api/openapi.ts"],
    rules: {
      "sonarjs/class-name": "off",
    },
  },
);
