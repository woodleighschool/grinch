import favicons from "favicons";
import { writeFile, mkdir } from "fs/promises";

const source = "src/assets/logo.png";

const config = {
  path: "/",
  icons: {
    favicons: true,
    appleIcon: false,
    appleStartup: false,
    android: false,
    windows: false,
    yandex: false,
  },
};

favicons(source, config)
  .then(async (res) => {
    await mkdir("public", { recursive: true });

    for (const image of res.images) {
      await writeFile(`public/${image.name}`, image.contents);
    }

    for (const file of res.files) {
      await writeFile(`public/${file.name}`, file.contents);
    }

    console.log("Favicons generated");
  })
  .catch(console.error);
