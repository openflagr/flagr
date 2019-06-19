<template>
  <div class="import-export-flags-container">
    <el-button type="primary" icon="el-icon-caret-bottom" v-on:click="exportFlags">Export all flags</el-button>
    <a ref="exportFile"/>

    <el-button class="import-btn" icon="el-icon-upload2" @click="importFlags">Import flags</el-button>
    <input type="file" hidden id="importFlags" @change="importFlagsChanged">
  </div>
</template>

<script>
import Axios from "axios";
import constants from "@/constants";

const { API_URL } = constants;

export default {
  name: "import-export-flags",
  props: ["flags"],
  methods: {
    async getFlag(flagId) {
      const { data } = await Axios.get(`${API_URL}/flags/${flagId}`);
      return data;
    },
    async getAllFlags() {
      const flagIds = this.flags.map(flag => flag.id);
      const getAllFlagsTasks = flagIds.map(this.getFlag);
      const flags = await Promise.all(getAllFlagsTasks);

      return flags;
    },
    async exportFlags() {
      const exportFileHref = this.$refs.exportFile;
      const flagsText = await this.getAllFlags();

      exportFileHref.setAttribute(
        "href",
        "data:text/plain;charset=utf-8," +
          encodeURIComponent(JSON.stringify(flagsText))
      );
      exportFileHref.setAttribute("download", "flags.json");
      exportFileHref.click();
    },
    onFileReaderLoaded(loadedFile) {
      const fileContent = loadedFile.target.result;

      try {
        const flags = JSON.parse(fileContent);
        debugger;
      } catch (err) {
        console.error("Failed to load flags from file", err);
      }
    },
    importFlags(e) {
      e.preventDefault();
      if (!window.FileReader) {
        console.info(
          "Reading files is not supported for that browser. Please try using anthor one."
        );
        return;
      }

      document.getElementById("importFlags").click();
    },
    importFlagsChanged(e) {
      const file = e.target.files[0];

      if (file.type !== "application/json") {
        console.info(
          "Import flags supports only JSON format, please upload a new file"
        );
        return;
      }

      const fileReader = new FileReader();
      fileReader.onload = this.onFileReaderLoaded;
      fileReader.readAsText(file);
    }
  }
};
</script>

<style lang="less" scoped>
.import-export-flags-container {
  .import-btn {
    margin-left: 10px;
  }
}
</style>
