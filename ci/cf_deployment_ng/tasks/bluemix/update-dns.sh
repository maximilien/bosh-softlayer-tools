#!/usr/bin/env bash
set -e

source bosh-softlayer-tools/ci/tasks/utils.sh
source /etc/profile.d/chruby.sh

check_param VCAP_PASSWORD

echo "[INFO] Install expect, ssh tool"
apt-get -y install expect ssh >/dev/null 2>&1
deployment_dir="${PWD}/deployment"
mkdir -p $deployment_dir

tar -zxvf director-artifacts/director_artifacts.tgz -C ${deployment_dir}
tar -zxvf cf-artifacts/cf_artifacts.tgz -C ${deployment_dir}
tar -zxvf run-utils/run-utils.tgz -C run-utils/

deploy_name=$(${deployment_dir}/bosh-cli* int ${deployment_dir}/cf-deploy-base.yml --path /name)
director_ip=$(awk '{print $1}' deployment/director-hosts)
domain1="${deploy_name}.bluebosh.com"
domain2="${deploy_name}.mybluebosh.com"
pg_password=$(${deployment_dir}/bosh-cli* int ${deployment_dir}/director-creds.yml --path /postgres_password)
ip_ha=$(grep ha_proxy ${deployment_dir}/deployed-vms|awk '{print $4}')

cat >run-utils/update_dns.sh<<EOF
cat >/tmp/update_dns.sql<<ENDSQL
DO \\\$\\\$
DECLARE new_id INTEGER;
BEGIN
    DELETE FROM domains WHERE domains.name IN  ('$domain1', '$domain2');
    DELETE FROM records WHERE records.name IN  ('$domain1', '*.$domain1', '$domain2', '*.$domain2');
    INSERT INTO domains(name, type) VALUES('${domain1}', 'NATIVE');
    SELECT domains.id INTO new_id from domains WHERE domains.name = '$domain1' ;
    INSERT INTO records(name, type, content, ttl, domain_id) VALUES('$domain1','SOA','localhost hostmaster@localhost 0 10800 604800 30', 300, new_id );
    INSERT INTO records(name, type, content, ttl, domain_id) VALUES('*.$domain1', 'A', '$ip_ha', 300, new_id);
    INSERT INTO domains(name, type) VALUES('$domain2', 'NATIVE');
    SELECT domains.id INTO new_id from domains where domains.name = '$domain2' ;
    INSERT INTO records(name, type, content, ttl, domain_id) VALUES('$domain2','SOA','localhost hostmaster@localhost 0 10800 604800 30', 300, new_id);
    INSERT INTO records(name, type, content, ttl, domain_id) VALUES('*.$domain2', 'A', '$ip_ha', 300,new_id);
END\\\$\\\$;
ENDSQL
/var/vcap/packages/postgres/bin/psql -U postgres -d powerdns -a -f /tmp/update_dns.sql
EOF

chmod +x run-utils/update_dns.sh
pushd run-utils
  ${deployment_dir}/bosh-cli* int ${deployment_dir}/director-creds.yml --path /jumpbox_ssh/private_key > ./jumpbox.key
  chmod 600 ./jumpbox.key
  ssh -o StrictHostKeyChecking=no jumpbox@${director_ip} -i ./jumpbox.key 'bash -s' < update_dns.sh
popd
