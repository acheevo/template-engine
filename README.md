# Template Engine

The CLI extracts reference projects into portable JSON templates and generates
new projects from those archives.

## Expo and React Native

Use the `react-native` type for the Acheevo `mobile-app` reference checkout.

```sh
template-engine extract ../mobile-app --type react-native -o mobile-app-template.json
template-engine generate mobile-app-template.json \
  --project-name "Care Notes" --github-repo example/care-notes \
  --output-dir ./care-notes
```

The generator updates the Expo display name, slug, scheme, npm package name,
and lockfile. Set the native bundle identifiers in the generated app before
linking it to an EAS project. Binary assets and hidden project configuration
travel with the archive, including `.nvmrc` and `.rnstorybook`. The extractor
skips development caches, generated native projects, and local credentials.

For a direct generate, run `template-engine new react-native "Care Notes" example/care-notes`
from a checkout alongside `mobile-app`. Existing reference configurations can
add the source with `template-engine config add react-native ../mobile-app`.
