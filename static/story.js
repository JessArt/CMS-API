$(function() {
  Vue.component('vue-input', {
    template: `
      <div class="form-group">
        <label>
          {{ label }}
        </label>
        <input class="form-control" v-bind:value="value" v-on:input="updateValue($event.target.value)" />
      </div>
    `,
    props: ['label', 'value'],
    methods: {
      updateValue(val) {
        this.$emit('input', val);
      }
    }
  });

  Vue.component('vue-textarea', {
    template: `
      <div class="form-group">
        <label>
          {{ label }}
        </label>
        <textarea class="form-control" v-bind:value="value" v-on:input="updateValue($event.target.value)" />
      </div>
    `,
    props: ['label', 'value'],
    methods: {
      updateValue(val) {
        this.$emit('input', val);
      }
    }
  });

  var photos = []
  superagent.get('/v2/api/images?type=photo')
    .end(function(err, res) {
      if (!err) {
        photos = res.body
      }
    })

  Vue.component('image-picker', {
    template: `
      <div class="image-picker">
        <div class="image-picker--content">
          <div class="image-picker--search">
            Write to filter:
            {{ search }}
            <input class="form-control" v-model="search" />
            <div v-on:click="close" class="image-picker--close">X</div>
          </div>
          <div v-for="photo in filteredPhotos" class="image-picker--item" v-on:click="selectImage(photo)">
            {{ photo.Title }}
            <img class="image-picker--image" :src="photo.SmallURL" />
          </div>
        </div>
      </div>
    `,
    data: function() {
      return {
        photos: photos,
        search: ''
      };
    },
    props: ['value'],
    methods: {
      selectImage: function(image) {
        this.$emit('choose', image);
      },
      close: function() {
        this.$emit('close');
      }
    },
    computed: {
      filteredPhotos: function() {
        return this.photos.filter(x => {
          return x.Tags.some(tag => tag.includes(this.search));
        });
      }
    }
  })

  Vue.component('story-image', {
    template: `
      <div class="story-image">
        <div class="">
          <template v-if="isVisible()">
            <div class="story-image--arrows">
              <div class="story-image--up" v-on:click="up"></div>
              <div class="story-image--down" v-on:click="down"></div>
            </div>
            <div>
              <vue-input label="Title" v-model="image.title" />
              <vue-textarea label="Description" v-model="image.description" />
            </div>
            <div>
              <img class="story-image--image" v-if="image.cover" :src="image.cover" />
            </div>
            <div>
              <template v-if="show">
                <image-picker v-on:close="closePicker" v-on:choose="selectImage" />
              </template>
              <button v-on:click="showPicker" class="btn">
                Choose a picture
              </button>
              <button v-on:click="remove" class="btn btn-danger">
                Delete
              </button>
            </div>
          </template>
          <template v-else>
            You want to remove "{{ image.title }}".
            <button class="btn btn-default" v-on:click="undo">
              Undo
            </button>
          </template>
        </div>
      </div>
    `,
    data: function() {
      return {
        show: false
      };
    },
    props: ['image'],
    methods: {
      up: function() {
        this.$emit('move', 'up', this.image);
      },
      down: function() {
        this.$emit('move', 'down', this.image);
      },
      isVisible: function() {
        return !this.image.remove;
      },
      isInvisible: function() {
        return !!this.image.remove;
      },
      showPicker: function(e) {
        if (e && e.preventDefault) {
          e.preventDefault();
        }
        this.show = true;
      },
      closePicker: function() {
        this.show = false;
      },
      selectImage: function(image) {
        this.show = false;
        this.image.imageId = image.ID;
        this.image.cover = image.BigURL;
        this.image.link = `https://jess.gallery/media/${image.ID}?type=${image.Type}`;
      },
      remove: function(e) {
        if (e && e.preventDefault) {
          e.preventDefault();
        }

        if (e.new) {
          const sure = confirm('Do you really want to delete this item? You cannot revert it.');

          if (sure) {
            this.$emit('remove', this.image);
          }
        } else {
          this.image.remove = true;
        }
      },
      undo: function(e) {
        if (e && e.preventDefault) {
          e.preventDefault();
        }

        this.image.remove = false;
      }
    }
  })

  window.DATA = window.DATA || {};
  Vue.component('story-form', {
    data: function(data) {
      return {
        id: DATA.ID || null,
        title: DATA.Title || '',
        subtitle: DATA.Subtitle || '',
        description: DATA.Description || '',
        cover: DATA.Cover || '',
        keywords: DATA.Keywords || '',
        metaTitle: DATA.MetaTitle || '',
        metaDescription: DATA.MetaDescription || '',
        images: (DATA.Images || []).map(image => Object.assign({}, image, {
          new: false,
          remove: false,
          id: image.ID,
          cover: image.Cover,
          title: image.Title,
          description: image.Description,
          imageId: image.ImageID,
          link: image.Link,
          sort: image.Sort
        })),
        status: null,
        chooseCover: false
      }
    },
    template: `
      <form>
        <vue-input label="Title" v-model="title" />
        <vue-input label="Subtitle" v-model="subtitle" />
        <vue-textarea label="Description" v-model="description" />
        <div style="cursor:pointer;text-decoration:underline;color:navy;display:inline-block;" v-on:click="showCover">
          Choose cover
        </div>
        <template v-if="chooseCover">
          <image-picker v-on:close="closePicker" v-on:choose="selectImage" />
        </template>
        <vue-input label="Cover" v-model="cover" />
        <vue-input label="Meta Keywords" v-model="keywords" />
        <vue-input label="Meta Title" v-model="metaTitle" />
        <vue-input label="Meta Description" v-model="metaDescription" />
        <ul id="example-1">
          <li v-for="image in images">
            <story-image v-on:move="move" :image="image" v-on:remove="remove" />
          </li>
        </ul>
        <button v-on:click="addImage" class="btn btn-primary">
          Add one more image
        </button>
        <div style="margin-top:20px;">
          <button v-on:click="submit" type="submit" class="btn btn-primary">
            Save
          </button>
          <h3 v-if="isLoading()">Saving...</h3>
        </div>
      </form>
    `,
    methods: {
      isLoading: function() {
        return this.status === 'loading';
      },
      addImage(e) {
        e.preventDefault();
        this.images.push({
          new: true,
          remove: false,
          imageId: null,
          id: null,
          title: '',
          description: '',
          cover: '',
          link: ''
        })
      },
      move: function(direction, image) {
        if (this.images.length > 1) {
          const index = this.images.findIndex(x => image === x);
          const newIndex = index + (direction === 'up' ? -1 : 1);

          if (newIndex >= 0 && newIndex < this.images.length) {
            this.images.splice(index, 1, this.images[newIndex]);
            this.images.splice(newIndex, 1, image);
          }
        }
      },
      remove: function(image) {
        this.images = this.images.filter(filteringImage => {
          return filteringImage !== image;
        });
      },
      showCover: function() {
        this.chooseCover = true;
      },
      closePicker: function() {
        this.chooseCover = false;
      },
      selectImage: function(image) {
        this.chooseCover = false;
        this.cover = `https:${image.BigURL}`;
      },
      submit: function(e) {
        if (e && e.preventDefault) {
          e.preventDefault();
        }
        this.status = 'loading';
        superagent
          .post('/story')
          .set('Accept', 'application/json')
          .set('Content-Type', 'application/json')
          .send({
            id: this.id,
            title: this.title,
            subtitle: this.subtitle,
            description: this.description,
            cover: this.cover,
            keywords: this.keywords,
            metaTitle: this.metaTitle,
            metaDescription: this.metaDescription,
            images: this.images.map((image, i) => Object.assign({}, image, {
              sort: i + 1
            }))
          })
          .end((err, res) => {
            if (err) {
              this.status = 'error';
            } else {
              this.status = 'success';
              location.reload();
            }
          })
      }
    }
  })

  var app = new Vue({
    el: '#app',
    data: {
      message: 'Hello Vue!'
    }
  });
})
