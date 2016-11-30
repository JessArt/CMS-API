var gulp = require('gulp');
var sass = require('gulp-sass');
var concat = require('gulp-concat');
var autoprefixer = require('gulp-autoprefixer');

gulp.task('styles', function() {
  gulp.src('src/client/styles/**/*.sass')
    .pipe(sass().on('error', sass.logError))
    .pipe(autoprefixer({
      browsers: ['last 2 versions'],
      cascade: false
    }))
    .pipe(concat('styles.css'))
    .pipe(gulp.dest('./static'));
});

gulp.task('js', function() {
  gulp.src('src/client/js/**/*.js')
    .pipe(gulp.dest('./static'));
});

gulp.task('default',function() {
  gulp.watch('src/client/**/*', ['styles', 'js']);
});
