<template>
  <div class="import-export-flags-container">
    <div v-if="importFlagsInProcess">
      Proccessing flags from files in work...
    </div>
    <el-button type="primary" icon="el-icon-caret-bottom" v-on:click="exportFlags">Export all flags</el-button>
    <a ref="exportFile"/>

    <el-button class="import-btn" icon="el-icon-upload2" @click="importFlags">Import flags</el-button>
    <input type="file" hidden id="importFlags" @change="importFlagsChanged">
  </div>
</template>

<script>
import Axios from 'axios';
import constants from '@/constants';

const { API_URL } = constants;

export default {
  name: "import-export-flags",
  props: ["flags", "loadFlags"],
  data () {
    return {
      importFlagsInProcess: false,
    }
  },
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
    createFlag(flag) {
      return Axios.post(`${API_URL}/flags`, {
        description: flag.description,
        key: flag.key
      });
    },
    updateFlagMetadata(flag, newFlagId) {
      const { dataRecordsEnabled, enabled, key, notes, entityType } = flag;

      return Axios.put(`${API_URL}/flags/${newFlagId}`, {
        dataRecordsEnabled,
        enabled,
        key,
        notes,
        entityType
      });
    },
    async createSegmentWithConstraints(segment, newFlagId, newVariants) {
      const {
        constraints,
        description,
        rolloutPercent,
        distributions
      } = segment;

      // 1. create segment
      const { data: createdSegment } = await Axios.post(
        `${API_URL}/flags/${newFlagId}/segments`,
        {
          description,
          rolloutPercent
        }
      );

      // 2. put constarints
      await Promise.all(
        constraints.map(({ id, ...propsToSend }) =>
          Axios.post(
            `${API_URL}/flags/${newFlagId}/segments/${
              createdSegment.id
            }/constraints`,
            {
              ...propsToSend // (property, operator, value)
            }
          )
        )
      );

      // since distributions points to non exsiting variant keys, we have to
      // update the data by newVariants
      const distributionsWithUpdatedVariantIds = distributions.map(
        ({ percent, variantKey }) => {
          const matchedVariant = newVariants.find(
            variant => variant.key === variantKey
          );

          if (!matchedVariant) {
            console.error("Could not find match for variant id", variantKey);
          }

          return {
            percent,
            variantKey,
            variantID: matchedVariant.id
          };
        }
      );

      // 3. put distribution
      await Axios.put(
        `${API_URL}/flags/${newFlagId}/segments/${
          createdSegment.id
        }/distributions`,
        {
          distributions: distributionsWithUpdatedVariantIds
        }
      );
    },
    async createSegments(segments, newFlagId, newVariants) {
      if (!segments) {
        return;
      }

      return Promise.all(
        segments.map(segment =>
          this.createSegmentWithConstraints(segment, newFlagId, newVariants)
        )
      );
    },
    async createVariants(variants, newFlagId) {
      const responses = await Promise.all(
        variants.map(({ key, attachment }) =>
          Axios.post(`${API_URL}/flags/${newFlagId}/variants`, {
            key,
            attachment
          })
        )
      );

      return responses.map(res => res.data);
    },
    async createFlagWithMetadata(flag) {
      const matchedFlag = this.flags.find(
        existingFlag =>
          existingFlag.id === flag.id ||
          existingFlag.description === flag.description
      );

      if (matchedFlag) {
        return;
      }

      const { segments, variants } = flag;

      try {
        const { data: newFlag } = await this.createFlag(flag);
        await this.updateFlagMetadata(flag, newFlag.id);
        const newVariants = await this.createVariants(
          variants,
          newFlag.id
        );
        await this.createSegments(segments, newFlag.id, newVariants); // TODO
      } catch (error) {
        console.error('Failed to create flag', error);
      }
    },
    async onFileReaderLoaded(loadedFile) {
      const fileContent = loadedFile.target.result;
      let flags;

      try {
        flags = JSON.parse(fileContent);
        this.importFlagsInProcess = true;
        await Promise.all(flags.map(this.createFlagWithMetadata));
        this.importFlagsInProcess = false;
        this.loadFlags()
      } catch (err) {
        if (!flags) {
          console.error("Failed to load flags from file", err);
        } else {
          console.error("Failed to update/create flag", err);
        }
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
