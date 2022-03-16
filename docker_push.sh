project_name="loyal-coyote-344317"
image_name="kv-store"

gcloud auth login
gcloud config set project $project_name

#todo add in steps to create the repo

gcloud auth configure-docker us-central1-docker.pkg.dev

docker tag $image_name us-central1-docker.pkg.dev/{$project_name}/kv-store-repo/{$image_name}
docker push us-central1-docker.pkg.dev/{$project_name}/kv-store-repo/{$image_name}