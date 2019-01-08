ARCH=$(shell uname -s | grep Darwin)
ifeq ($(ARCH),Darwin)
	OPTS=-it
else
	OPTS=-i
endif

IS_GAE_EXIST=$(shell gcloud app services list && echo 1 || echo 0)

readme:
	@echo 'Support Sub Commands:'
	@echo ''
	@echo 'Set porject ID'
	@echo ''
	@echo '    $$ make set PROJECT=<your_gcp_project_id>'
	@echo ''
	@echo 'Set YAMLs'
	@echo ''
	@echo '    $$ make yaml SERVICE=<your_service_name>'
	@echo ''

set:
	@echo ''
	@echo 'Set porject ID'
	@echo ''

	gcloud config set project $(PROJECT)

yaml:
	@echo ''
	@echo 'Set yamls'
	@echo 'If you get `ERROR: Could not find Application`, it mean you DO NOT need to change serivce'
	@echo ''

 ifeq ($(IS_GAE_EXIST),0)
	echo 'nothing to do'
 else
	sed $(OPTS) 's/default/'$(SERVICE)'/' ./app/app.yaml ./app/cron.yaml ./app/queue.yaml
 endif
