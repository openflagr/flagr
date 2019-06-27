<template>
  <div class="import-export-flags-container">
    <div v-if="importFlagsInProcess">Proccessing flags from files in work...</div>
    <el-button
      type="primary"
      icon="el-icon-caret-bottom"
      v-on:click="exportFlags"
      :disabled="this.importFlagsInProcess"
    >Export all flags</el-button>
    <a ref="exportFile"/>

    <el-button
      class="import-btn"
      icon="el-icon-upload2"
      @click="importFlags"
      :disabled="this.importFlagsInProcess"
    >Import flags</el-button>
    <input type="file" hidden id="importFlags" @change="importFlagsChanged">
    <el-row>
      <el-progress
        class="import-flags-progeress"
        v-if="this.importFlagsInProcess"
        type="circle"
        :percentage="progressPercentage"
      ></el-progress>
    </el-row>
  </div>
</template>

<script>
import Axios from "axios";
import constants from "@/constants";

const { API_URL } = constants;

export default {
  name: "import-export-flags",
  props: ["flags", "loadFlags"],
  data() {
    return {
      importFlagsInProcess: false,
      numOfFlagsToProcess: 0,
      finishedImportFlags: 0
    };
  },
  computed: {
    progressPercentage() {
      return (this.finishedImportFlags / (this.numOfFlagsToProcess || 1)) * 100;
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
    updateDistributions(variants, flagId, segmentId, distributions) {
      // since distributions points to non exsiting variant keys, we have to
      // update the data by newVariants
      const distributionsWithUpdatedVariantIds = distributions.map(
        ({ percent, variantKey }) => {
          const matchedVariant = variants.find(
            variant => variant.key === variantKey
          );

          if (!matchedVariant) {
            this.$message({
              message: `Could not find match for variant key=${variantKey}`,
              type: "error"
            });
          }

          return {
            percent,
            variantKey,
            variantID: matchedVariant.id
          };
        }
      );

      // 3. put distribution
      return Axios.put(
        `${API_URL}/flags/${flagId}/segments/${segmentId}/distributions`,
        {
          distributions: distributionsWithUpdatedVariantIds
        }
      );
    },
    updateConstarints(flagId, segmentId, constraints) {
      return Promise.all(
        // eslint-disable-next-line
        constraints.map(({ id, ...propsToSend }) =>
          Axios.post(
            `${API_URL}/flags/${flagId}/segments/${segmentId}/constraints`,
            {
              ...propsToSend // (property, operator, value)
            }
          )
        )
      );
    },
    createSegment(flagId, updateData) {
      return Axios.post(`${API_URL}/flags/${flagId}/segments`, updateData);
    },
    async createSegmentWithConstraints(segment, newFlagId, newVariants) {
      const {
        constraints,
        description,
        rolloutPercent,
        distributions
      } = segment;

      const { data: createdSegment } = await this.createSegment(newFlagId, {
        description,
        rolloutPercent
      });

      await this.updateConstarints(newFlagId, createdSegment.id, constraints);
      await this.updateDistributions(
        newVariants,
        newFlagId,
        createdSegment.id,
        distributions
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

      const { data: newFlag } = await this.createFlag(flag);
      await this.updateFlagMetadata(flag, newFlag.id);
      const newVariants = await this.createVariants(variants, newFlag.id);
      await this.createSegments(segments, newFlag.id, newVariants);
      this.finishedImportFlags++;
    },
    async onFileReaderLoaded(loadedFile) {
      const fileContent = loadedFile.target.result;
      let flags;

      try {
        flags = JSON.parse(fileContent);
        this.numOfFlagsToProcess = flags.length;
        this.importFlagsInProcess = true;
        await Promise.all(flags.map(this.createFlagWithMetadata));

        this.$message({
          message: `Finished import flags successfully!`,
          type: "success"
        });

        this.loadFlags();
      } catch (err) {
        if (!flags) {
          this.$message({
            message: `Failed to load flags from file`,
            type: "error"
          });
        } else {
          this.$message({
            message: `Failed to update/create flag`,
            type: "error"
          });
        }
      } finally {
        this.importFlagsInProcess = false;
      }
    },
    importFlags(e) {
      e.preventDefault();
      if (!window.FileReader) {
        return this.$message({
          message: `Reading files is not supported for that browser. Please try using anthor one.`,
          type: "error"
        });
      }

      document.getElementById("importFlags").click();
    },
    importFlagsChanged(e) {
      const file = e.target.files[0];

      if (file.type !== "application/json") {
        return this.$message({
          message: `Import flags supports only JSON format, please upload a new file`,
          type: "error"
        });
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
  .import-flags-progeress {
    margin: 30px 45%;
  }
}
</style>
