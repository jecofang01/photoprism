<template>
  <div class="p-tab p-tab-photo-people">
    <v-container grid-list-xs fluid class="pa-2 p-faces">
      <v-card v-if="markers.length === 0" class="no-results secondary-light lighten-1 ma-1" flat>
        <v-card-title primary-title>
          <div>
            <h3 class="title ma-0 pa-0">
              <translate>Couldn't find any people</translate>
            </h3>
            <p class="mt-4 mb-0 pa-0">
              <translate>To automatically detect faces, please re-index your library and wait until indexing has been completed.</translate>
            </p>
          </div>
        </v-card-title>
      </v-card>
      <v-layout row wrap class="search-results face-results cards-view">
        <v-flex
            v-for="(marker, index) in markers"
            :key="index"
            xs6 sm4 md3 lg2 xl1 d-flex
        >
          <v-card tile
                  :data-id="marker.UID"
                  style="user-select: none"
                  :class="{invalid: marker.Invalid}"
                  class="result accent lighten-3">
            <div class="card-background accent lighten-3"></div>
            <canvas :id="'face-' + marker.UID" :key="marker.UID" width="300" height="300" style="width: 100%"
                    class="v-responsive v-image accent lighten-2"></canvas>

            <v-card-actions class="card-details pa-0">
              <v-layout v-if="marker.Review || marker.Invalid" row wrap align-center>
                <v-flex xs6 class="text-xs-center pa-0">
                  <v-btn color="transparent" :disabled="busy"
                         large depressed block :round="false"
                         class="action-archive text-xs-center"
                         :title="$gettext('Reject')" @click.stop="reject(marker)">
                    <v-icon dark>clear</v-icon>
                  </v-btn>
                </v-flex>
                <v-flex xs6 class="text-xs-center pa-0">
                  <v-btn color="transparent" :disabled="busy"
                         large depressed block :round="false"
                         class="action-approve text-xs-center"
                         :title="$gettext('Approve')" @click.stop="approve(marker)">
                    <v-icon dark>check</v-icon>
                  </v-btn>
                </v-flex>
              </v-layout>
              <v-layout v-else row wrap align-center>
                <v-flex xs12 class="text-xs-left pa-0">
                  <v-text-field
                      v-model="marker.Name"
                      :rules="[textRule]"
                      :disabled="busy"
                      browser-autocomplete="off"
                      class="input-name pa-0 ma-0"
                      hide-details
                      single-line
                      solo-inverted
                      clearable
                      @click:clear="clearSubject(marker)"
                      @change="rename(marker)"
                      @keyup.enter.native="rename(marker)"
                  ></v-text-field>
                </v-flex>
              </v-layout>
            </v-card-actions>
          </v-card>
        </v-flex>
      </v-layout>
    </v-container>
  </div>
</template>

<script>

export default {
  name: 'PTabPhotoPeople',
  props: {
    model: Object,
    uid: String,
  },
  data() {
    return {
      busy: false,
      markers: this.model.getMarkers(true),
      imageUrl: this.model.thumbnailUrl("fit_720"),
      disabled: !this.$config.feature("edit"),
      config: this.$config.values,
      readonly: this.$config.get("readonly"),
      textRule: (v) => {
        if (!v || !v.length) {
          return this.$gettext("Name");
        }

        return v.length <= this.$config.get('clip') || this.$gettext("Text too long");
      },
    };
  },
  mounted() {
    this.markers.forEach((m) => {
      const canvas = document.getElementById('face-' + m.UID);

      let ctx = canvas.getContext('2d');
      let img = new Image();

      img.onload = function () {
        const w = Math.round(m.W * img.width);
        const h = Math.round(m.H * img.height);
        const s = w > h ? w : h;
        const x = Math.round(m.X * img.width);
        const y = Math.round(m.Y * img.height);

        /*
        console.log("Crop X", m.X, x);
        console.log("Crop Y", m.Y, y);
        console.log("Crop W", m.W, w);
        console.log("Crop H", m.H, h);
        console.log("Image WxH", img.width, img.height);
        */

        ctx.drawImage(img, x, y, s, s, 0, 0, 300, 300);
      };

      if (m.W < 0.07) {
        // TODO: Not all users have thumbs with this resolution.
        img.src = this.model.thumbnailUrl("fit_7680");
      } else if (m.W < 0.1) {
        // TODO: Not all users have thumbs with this resolution.
        img.src = this.model.thumbnailUrl("fit_2048");
      } else if (m.W < 0.15) {
        // TODO: Not all users have thumbs with this resolution.
        img.src = this.model.thumbnailUrl("fit_1280");
      } else {
        img.src = this.imageUrl;
      }
    });
  },
  methods: {
    refresh() {
    },
    reject(marker) {
      this.busy = true;
      marker.reject().finally(() => this.busy = false);
    },
    approve(marker) {
      this.busy = true;
      marker.approve().finally(() => this.busy = false);
    },
    clearSubject(marker) {
      this.busy = true;
      marker.clearSubject(marker).finally(() => this.busy = false);
    },
    rename(marker) {
      this.busy = true;
      marker.rename().finally(() => this.busy = false);
    },
  },
};
</script>
