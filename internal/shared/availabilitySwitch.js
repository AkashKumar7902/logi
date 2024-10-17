// src/components/Shared/AvailabilitySwitch.js

import React from 'react';
import Switch from 'react-switch';
import PropTypes from 'prop-types';

const AvailabilitySwitch = ({ isOnline, handleToggle }) => {
  return (
    <div className="flex items-center mb-4">
      <span className="mr-2 text-gray-700">Offline</span>
      <Switch
        onChange={handleToggle}
        checked={isOnline}
        onColor="#86d3ff"
        onHandleColor="#2693e6"
        handleDiameter={30}
        uncheckedIcon={false}
        checkedIcon={false}
        boxShadow="0px 1px 5px rgba(0, 0, 0, 0.6)"
        activeBoxShadow="0px 0px 1px 10px rgba(0, 0, 0, 0.2)"
        height={20}
        width={48}
        className="react-switch"
        id="material-switch"
      />
      <span className="ml-2 text-gray-700">Online</span>
    </div>
  );
};

AvailabilitySwitch.propTypes = {
  isOnline: PropTypes.bool.isRequired,
  handleToggle: PropTypes.func.isRequired,
};

export default AvailabilitySwitch;
