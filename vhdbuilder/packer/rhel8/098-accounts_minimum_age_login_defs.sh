#!/usr/bin/env bash
set -e
(>&2 echo "Remediating rule 98/364: 'accounts_minimum_age_login_defs'")
# Remediation is applicable only in certain platforms
if rpm --quiet -q shadow-utils; then


var_accounts_minimum_age_login_defs="1"



grep -q ^PASS_MIN_DAYS /etc/login.defs && \
  sed -i "s/PASS_MIN_DAYS.*/PASS_MIN_DAYS     $var_accounts_minimum_age_login_defs/g" /etc/login.defs
if ! [ $? -eq 0 ]; then
    echo "PASS_MIN_DAYS      $var_accounts_minimum_age_login_defs" >> /etc/login.defs
fi

else
    >&2 echo 'Remediation is not applicable, nothing was done'
fi
